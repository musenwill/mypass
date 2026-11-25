package util

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/musenwill/mypass/errs"
)

func TestRecordKey(t *testing.T) {
	key := RecordKey([]byte("token"), "title")

	if act, exp := len(key), KeyLen; act != exp {
		t.Errorf("got %v key len expected %v", act, exp)
	}
	if bytes.Equal(key, RecordKey([]byte("token"), "another title")) {
		t.Error("got same key expected keys of two titles differ")
	}
	if bytes.Equal(key, RecordKey([]byte("another token"), "title")) {
		t.Error("got same key expected keys of two tokens differ")
	}
}

func TestEncryptAndDecrypt(t *testing.T) {
	origin := "world"

	api := NewCrypto(RecordKey([]byte("token"), "title"), []byte("title"))
	cipherText, err := api.Encrypt([]byte(origin))
	if err != nil {
		t.Error(err)
	}
	plainText, err := api.Decrypt(cipherText)
	if err != nil {
		t.Error(err)
	}

	if act, exp := string(plainText), origin; act != exp {
		t.Errorf("got %v decrypted expected %v", act, exp)
	}
}

func TestEncryptUsesFreshNonce(t *testing.T) {
	api := NewCrypto(RecordKey([]byte("token"), "title"), []byte("title"))

	first, err := api.Encrypt([]byte("world"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := api.Encrypt([]byte("world"))
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(first, second) {
		t.Error("got same ciphertext expected a fresh nonce on every encryption")
	}
}

func TestDecryptFailsOnChangedCiphertext(t *testing.T) {
	api := NewCrypto(RecordKey([]byte("token"), "title"), []byte("title"))
	cipherText, err := api.Encrypt([]byte("world"))
	if err != nil {
		t.Fatal(err)
	}

	// change every byte one by one, each one must break the decryption
	for i := range cipherText {
		changed := append([]byte(nil), cipherText...)
		changed[i] ^= 0x01

		if _, err := api.Decrypt(changed); err != errs.DecryptError {
			t.Errorf("got %v expected decrypt error after changing byte %v", err, i)
		}
	}

	if _, err := api.Decrypt(cipherText[:len(cipherText)-1]); err != errs.DecryptError {
		t.Errorf("got %v expected decrypt error on truncated ciphertext", err)
	}
}

func TestDecryptFailsOnWrongKeyOrAad(t *testing.T) {
	key := RecordKey([]byte("token"), "title")
	cipherText, err := NewCrypto(key, []byte("title")).Encrypt([]byte("world"))
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]CryptoApi{
		"wrong key":   NewCrypto(RecordKey([]byte("another token"), "title"), []byte("title")),
		"wrong aad":   NewCrypto(key, []byte("another title")),
		"short input": NewCrypto(key, []byte("title")),
	}
	for name, api := range cases {
		content := cipherText
		if name == "short input" {
			content = cipherText[:NonceLen-1]
		}
		if _, err := api.Decrypt(content); err != errs.DecryptError {
			t.Errorf("got %v expected decrypt error on %v", err, name)
		}
	}
}

func TestLegacyEncryptAndDecrypt(t *testing.T) {
	origin := "world"

	api := NewLegacyHMacCrypto([]byte("hello"), []byte(origin))
	cipherText, err := api.Encrypt([]byte(origin))
	if err != nil {
		t.Error(err)
	}
	plainText, err := api.Decrypt(cipherText)
	if err != nil {
		t.Error(err)
	}

	if act, exp := string(plainText), origin; act != exp {
		t.Errorf("got %v decrypted expected %v", act, exp)
	}
}

// The vectors are produced by the v1 code itself, the legacy crypto must keep
// reading them, and encrypting the same way, to upgrade a v1 store.
const (
	// pincode "pin", plaintext "tencent,qq\n"
	legacyStoreVector = "r7sDbfkHLPEvs4+G8GcvGA=="
	// title "qq", token "token", plaintext "admin"
	legacyRecordVector = "H2PNLhOPZtfFiFW400Jklg=="
)

func TestLegacyVectors(t *testing.T) {
	cases := map[string]struct {
		crypto CryptoApi
		vector string
		origin string
	}{
		"store":  {NewLegacyCrypto([]byte("pin")), legacyStoreVector, "tencent,qq\n"},
		"record": {NewLegacyHMacCrypto([]byte("qq"), []byte("token")), legacyRecordVector, "admin"},
	}

	for name, c := range cases {
		cipherText, err := base64.StdEncoding.DecodeString(c.vector)
		if err != nil {
			t.Fatal(err)
		}

		plainText, err := c.crypto.Decrypt(cipherText)
		if err != nil {
			t.Errorf("got %v expected no error on %v", err, name)
			continue
		}
		if act, exp := string(plainText), c.origin; act != exp {
			t.Errorf("got %v decrypted expected %v on %v", act, exp, name)
		}

		again, err := c.crypto.Encrypt([]byte(c.origin))
		if err != nil {
			t.Error(err)
			continue
		}
		if act, exp := base64.StdEncoding.EncodeToString(again), c.vector; act != exp {
			t.Errorf("got %v expected %v, the v1 encryption is deterministic", act, exp)
		}
	}
}

// About 0.2 s/op warm and 0.3 s in a fresh process for the default params, on
// an i5-10400. A list command derives once, a `get` twice.
func BenchmarkKdf(b *testing.B) {
	for _, p := range []KdfParams{
		{Name: kdfArgon2id, Memory: 19456, Time: 2, Threads: 1}, // owasp minimum
		DefaultKdfParams,
		{Name: kdfArgon2id, Memory: 262144, Time: 1, Threads: 1},
	} {
		b.Run(fmt.Sprintf("m%dt%dp%d", p.Memory, p.Time, p.Threads), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := p.Derive([]byte("pin"), []byte("salt")); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
