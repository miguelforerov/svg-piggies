package productimages

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"testing"
)

func TestValidateImageBytes(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	img.Set(0, 0, color.NRGBA{R: 255, A: 255})
	var pngData, jpegData bytes.Buffer
	if err := png.Encode(&pngData, img); err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(&jpegData, img, nil); err != nil {
		t.Fatal(err)
	}
	// The 442-byte WebP fixture is from golang.org/x/image v0.45.0 testdata
	// (gopher-doc.1bpp.lossless.webp, BSD-3-Clause).
	webpData, err := os.ReadFile("testdata/valid.webp")
	if err != nil {
		t.Fatal(err)
	}
	for _, sample := range []struct {
		mime string
		data []byte
	}{
		{"image/png", pngData.Bytes()},
		{"image/jpeg", jpegData.Bytes()},
		{"image/webp", webpData},
	} {
		t.Run(sample.mime, func(t *testing.T) {
			width, height, err := validateImageBytes(sample.data, sample.mime)
			if err != nil || width <= 0 || height <= 0 {
				t.Fatalf("valid image rejected: %v", err)
			}
			if _, _, err := validateImageBytes(sample.data[:len(sample.data)/2], sample.mime); err == nil {
				t.Fatal("truncated image accepted")
			}
			corrupt := bytes.Repeat([]byte{0}, len(sample.data))
			copy(corrupt[:min(16, len(corrupt))], sample.data[:min(16, len(sample.data))])
			if _, _, err := validateImageBytes(corrupt, sample.mime); err == nil {
				t.Fatal("image with plausible header and corrupt body accepted")
			}
		})
	}
	if _, _, err := validateImageBytes(pngData.Bytes(), "image/avif"); err == nil {
		t.Fatal("mismatched MIME accepted")
	}
	// Even a plausible AVIF header must be rejected before any AVIF decoder runs.
	if _, _, err := validateImageBytes([]byte("\x00\x00\x00\x10ftypavif\x00\x00\x00\x00"), "image/avif"); err == nil || err.Error() != "unsupported image format" {
		t.Fatalf("AVIF rejection = %v", err)
	}
	hugePNG := append([]byte(nil), pngData.Bytes()...)
	binary.BigEndian.PutUint32(hugePNG[16:20], 8192)
	binary.BigEndian.PutUint32(hugePNG[20:24], 8192)
	binary.BigEndian.PutUint32(hugePNG[29:33], crc32.ChecksumIEEE(hugePNG[12:29]))
	if _, _, err := validateImageBytes(hugePNG, "image/png"); err == nil {
		t.Fatal("compressed image with excessive dimensions accepted")
	}
}
