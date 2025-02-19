package cryptocodec

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

type CryptoType uint32

const (
	AESGCM CryptoType = iota
	XOR
	Plain
	Unknown
)

var cryptoTypeNameMap = map[CryptoType]string{
	Unknown: "",
	Plain:   "plain",
	XOR:     "xor",
	AESGCM:  "aesgcm",
}

func (c CryptoType) String() string {
	if name, ok := cryptoTypeNameMap[c]; ok {
		return name
	}
	return fmt.Sprintf("unknown crypto type: %d", c)
}

var cryptoMap = map[CryptoType]Crypto{
	Plain:  plainCrypto,
	XOR:    xorCrypto,
	AESGCM: aesgcmCrypto,
}

func (c CryptoType) Crypto() Crypto {
	if crypto, ok := cryptoMap[c]; ok {
		return crypto
	}
	return nil
}

func (c CryptoType) Encrypt(plaintext, key []byte) ([]byte, error) {
	if crypto := c.Crypto(); crypto != nil {
		return crypto.Encrypt(plaintext, key)
	}
	return nil, fmt.Errorf("unknown crypto type: %d", c)
}

func (c CryptoType) Decrypt(ciphertext, key []byte) ([]byte, error) {
	if crypto := c.Crypto(); crypto != nil {
		return crypto.Decrypt(ciphertext, key)
	}
	return nil, fmt.Errorf("unknown crypto type: %d", c)
}

type Crypto interface {
	fmt.Stringer
	Encrypt(plaintext, key []byte) ([]byte, error)
	Decrypt(ciphertext, key []byte) ([]byte, error)
}

type CryptoEntity struct {
	Key    []byte
	Crypto Crypto
}

func NewCryptoEntity(key []byte, crypto Crypto) *CryptoEntity {
	return &CryptoEntity{
		Key:    key,
		Crypto: crypto,
	}
}

func (c *CryptoEntity) Encrypt(data []byte) ([]byte, error) {
	return c.Crypto.Encrypt(data, c.Key)
}

func (c *CryptoEntity) Decrypt(data []byte) ([]byte, error) {
	return c.Crypto.Decrypt(data, c.Key)
}

type PlainCrypto struct{}

var plainCrypto = &PlainCrypto{}

func (*PlainCrypto) String() string {
	return Plain.String()
}

func (*PlainCrypto) Encrypt(plaintext, key []byte) ([]byte, error) {
	return plaintext, nil
}

func (*PlainCrypto) Decrypt(ciphertext, key []byte) ([]byte, error) {
	return ciphertext, nil
}

type XORCrypto struct{}

var xorCrypto = &XORCrypto{}

func (*XORCrypto) String() string {
	return XOR.String()
}

func (*XORCrypto) Encrypt(plaintext, key []byte) ([]byte, error) {
	ciphertext := make([]byte, len(plaintext))
	for i := range plaintext {
		ciphertext[i] = plaintext[i] ^ key[i%len(key)]
	}
	return ciphertext, nil
}

func (*XORCrypto) Decrypt(ciphertext, key []byte) ([]byte, error) {
	plaintext := make([]byte, len(ciphertext))
	for i := range ciphertext {
		plaintext[i] = ciphertext[i] ^ key[i%len(key)]
	}
	return plaintext, nil
}

type AESGCMCrypto struct{}

var aesgcmCrypto = &AESGCMCrypto{}

func (*AESGCMCrypto) String() string {
	return AESGCM.String()
}

func (aesgcm *AESGCMCrypto) Encrypt(plaintext, key []byte) ([]byte, error) {
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// test code:
	// nonce := []byte("123456789012")
	return aesgcm.EncryptWithNonce(plaintext, key, nonce)
}

func (*AESGCMCrypto) EncryptWithNonce(plaintext, key []byte, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ad := aead.Seal(nil, nonce, nonce, nonce)
	ciphertext := aead.Seal(nil, nonce, plaintext, ad)

	return append(nonce, ciphertext...), nil
}

func (*AESGCMCrypto) Decrypt(ciphertext, key []byte) ([]byte, error) {
	if len(ciphertext) < 12 {
		return nil, fmt.Errorf("invalid ciphertext")
	}

	nonce := ciphertext[:12]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ad := aead.Seal(nil, nonce, nonce, nonce)
	plaintext, err := aead.Open(nil, nonce, ciphertext[12:], ad)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
