package cryptocodec_test

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/aura-studio/nano/codec/cryptocodec"
)

func TestCrypto(t *testing.T) {
	var key = make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Error(err)
	}

	for i := 1; i < 100; i++ {
		var plaintext = make([]byte, i)
		if _, err := rand.Read(plaintext); err != nil {
			t.Error(err)
		}

		var aesgcm = &cryptocodec.AESGCM{}
		data, err := aesgcm.Encrypt(plaintext, key)
		if err != nil {
			t.Error(err)
		}

		t.Log(base64.StdEncoding.EncodeToString(data))

		plaintextNew, err := aesgcm.Decrypt(data, key)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(plaintext, plaintextNew) {
			t.Error("plaintext not equal")
		}
	}
}
