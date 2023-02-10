package cryptocodec

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

type Method int

const (
	AES128GCM Method = iota
	AES256GCM
)

var methodNameMap = map[Method]string{
	AES128GCM: "aes-128-gcm",
	AES256GCM: "aes-256-gcm",
}

func (m Method) String() string {
	if name, ok := methodNameMap[m]; ok {
		return name
	}
	return fmt.Sprintf("unknown method: %d", m)
}

type Crypto struct {
	Method Method
	Key    []byte
}

func NewCrypto(method Method, key []byte) (*Crypto, error) {
	c := &Crypto{
		Method: method,
		Key:    key,
	}

	return c, nil
}

func (c *Crypto) Encode(data []byte) ([]byte, error) {
	switch c.Method {
	case AES128GCM, AES256GCM:
		return aesgcm.Encrypt(data, c.Key)
	default:
		return nil, fmt.Errorf("unknown method: %s", c.Method)
	}
}

func (c *Crypto) Decode(data []byte) ([]byte, error) {
	switch c.Method {
	case AES128GCM, AES256GCM:
		return aesgcm.Decrypt(data, c.Key)
	default:
		return nil, fmt.Errorf("unknown method: %s", c.Method)
	}
}

type AESGCM struct{}

var aesgcm = &AESGCM{}

func (*AESGCM) Encrypt(plaintext, key []byte) ([]byte, error) {
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// test code:
	// nonce := []byte("123456789012")
	return aesgcm.EncryptWithNonce(plaintext, key, nonce)
}

func (*AESGCM) EncryptWithNonce(plaintext, key []byte, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ad := aesgcm.Seal(nil, nonce, nonce, nonce)
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, ad)

	return append(nonce, ciphertext...), nil
}

func (*AESGCM) Decrypt(ciphertext, key []byte) ([]byte, error) {
	if len(ciphertext) < 12 {
		return nil, fmt.Errorf("invalid ciphertext")
	}

	nonce := ciphertext[:12]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ad := aesgcm.Seal(nil, nonce, nonce, nonce)
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext[12:], ad)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
