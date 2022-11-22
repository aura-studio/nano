package aes

import (
	"crypto/aes"
)

const BlockSize = aes.BlockSize

func Encode(key, plain []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	crypted := make([]byte, len(plain))
	for i := 0; i < len(plain); i += aes.BlockSize {
		c.Encrypt(crypted[i:i+aes.BlockSize], plain[i:i+aes.BlockSize])
	}
	return crypted, nil
}

func Decode(key, crypted []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plain := make([]byte, len(crypted))
	for i := 0; i < len(crypted); i += aes.BlockSize {
		c.Decrypt(plain[i:i+aes.BlockSize], crypted[i:i+aes.BlockSize])
	}
	return plain, nil
}
