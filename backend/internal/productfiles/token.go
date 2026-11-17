package productfiles

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"strings"

	"github.com/google/uuid"
)

const maxTokenBytes = 4096

type claims struct {
	KeyVersion    uint32
	ProductID     uuid.UUID
	Nonce         [32]byte
	FileSizeBytes int64
	ExpiresAt     int64
	FileKey       string
	FileName      string
	Format        string
	ContentType   string
}

// canonicalClaims encodes PF1, key version (u32), product UUID (16 bytes),
// nonce (32 bytes), size and Unix expiry (i64), then four u16-length-prefixed
// UTF-8 strings: object key, file name, format, and content type. Integers are
// big-endian. The same bytes are authenticated and fingerprinted with distinct keys.
func canonicalClaims(c claims) []byte {
	var b bytes.Buffer
	b.WriteString("PF1")
	_ = binary.Write(&b, binary.BigEndian, c.KeyVersion)
	b.Write(c.ProductID[:])
	b.Write(c.Nonce[:])
	_ = binary.Write(&b, binary.BigEndian, c.FileSizeBytes)
	_ = binary.Write(&b, binary.BigEndian, c.ExpiresAt)
	for _, value := range []string{c.FileKey, c.FileName, c.Format, c.ContentType} {
		_ = binary.Write(&b, binary.BigEndian, uint16(len(value)))
		b.WriteString(value)
	}
	return b.Bytes()
}

func parseClaims(payload []byte) (claims, error) {
	var c claims
	if len(payload) < 83 || len(payload) > maxTokenBytes || string(payload[:3]) != "PF1" {
		return c, ErrInvalidUpload
	}
	r := bytes.NewReader(payload[3:])
	if binary.Read(r, binary.BigEndian, &c.KeyVersion) != nil ||
		readBytes(r, c.ProductID[:]) != nil || readBytes(r, c.Nonce[:]) != nil ||
		binary.Read(r, binary.BigEndian, &c.FileSizeBytes) != nil ||
		binary.Read(r, binary.BigEndian, &c.ExpiresAt) != nil {
		return claims{}, ErrInvalidUpload
	}
	values := []*string{&c.FileKey, &c.FileName, &c.Format, &c.ContentType}
	for _, value := range values {
		var length uint16
		if binary.Read(r, binary.BigEndian, &length) != nil || length == 0 || int(length) > r.Len() {
			return claims{}, ErrInvalidUpload
		}
		data := make([]byte, length)
		if readBytes(r, data) != nil {
			return claims{}, ErrInvalidUpload
		}
		*value = string(data)
	}
	if r.Len() != 0 || !bytes.Equal(canonicalClaims(c), payload) || c.KeyVersion == 0 {
		return claims{}, ErrInvalidUpload
	}
	return c, nil
}

func readBytes(r io.Reader, data []byte) error {
	_, err := io.ReadFull(r, data)
	return err
}

func signClaims(c claims, key []byte) string {
	payload := canonicalClaims(c)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("svg-piggies:product-file-upload:v1:\x00"))
	mac.Write(payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func verifyClaims(token string, keys map[int]KeyPair) (claims, []byte, error) {
	var c claims
	if len(token) == 0 || len(token) > maxTokenBytes*2 || strings.Count(token, ".") != 1 {
		return c, nil, ErrInvalidUpload
	}
	parts := strings.SplitN(token, ".", 2)
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return c, nil, ErrInvalidUpload
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(signature) != sha256.Size {
		return c, nil, ErrInvalidUpload
	}
	c, err = parseClaims(payload)
	if err != nil {
		return c, nil, err
	}
	key, ok := keys[int(c.KeyVersion)]
	if !ok || len(key.Signing) != 32 || len(key.Fingerprint) != 32 {
		return claims{}, nil, ErrInvalidUpload
	}
	mac := hmac.New(sha256.New, key.Signing)
	mac.Write([]byte("svg-piggies:product-file-upload:v1:\x00"))
	mac.Write(payload)
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return claims{}, nil, ErrInvalidUpload
	}
	return c, payload, nil
}

func fingerprint(payload []byte, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("svg-piggies:product-file-confirmation:v1:\x00"))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// A consumed token's exact bytes can be recognized after its HMAC key is retired.
// The digest is persisted only after the token has been authenticated successfully.
func tokenDigest(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func validateKeys(keys map[int]KeyPair, active int) error {
	if active <= 0 || active > (1<<31-1) || len(keys) == 0 {
		return ErrUploadUnavailable
	}
	for version, key := range keys {
		if version <= 0 || version > (1<<31-1) || len(key.Signing) != 32 || len(key.Fingerprint) != 32 || hmac.Equal(key.Signing, key.Fingerprint) {
			return errors.New("product file key versions require distinct 32-byte signing and fingerprint keys")
		}
	}
	if _, ok := keys[active]; !ok {
		return ErrUploadUnavailable
	}
	return nil
}
