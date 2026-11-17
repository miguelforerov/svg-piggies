package config

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestProductFileKeyConfiguration(t *testing.T) {
	a := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32)))
	b := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("b", 32)))
	c := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("c", 32)))
	d := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("d", 32)))
	raw := `{"1":{"signing":"` + a + `","fingerprint":"` + b + `"},"2":{"signing":"` + c + `","fingerprint":"` + d + `"}}`
	keys, active, err := parseProductFileKeys(func(name string) string {
		if name == "PRODUCT_FILE_KEYS_JSON" {
			return raw
		}
		if name == "PRODUCT_FILE_ACTIVE_KEY_VERSION" {
			return "2"
		}
		return ""
	})
	if err != nil || active != 2 || len(keys) != 2 || len(keys[1].Signing) != 32 {
		t.Fatalf("key rotation config: %d %v %v", active, keys, err)
	}
	_, _, err = parseProductFileKeys(func(name string) string {
		if name == "PRODUCT_FILE_KEYS_JSON" {
			return `{"1":{"signing":"` + a + `","fingerprint":"` + a + `"}}`
		}
		if name == "PRODUCT_FILE_ACTIVE_KEY_VERSION" {
			return "1"
		}
		return ""
	})
	if err == nil {
		t.Fatal("identical keys accepted")
	}
}
