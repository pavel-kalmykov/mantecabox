// Package aes is a reference implementacion from https://gist.github.com/manishtpatel/8222606.
// In this package we process errors with panic because this cipher must be always working properly for this program
package utilities

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

type (
	AesCTRCipher interface {
		Encrypt(plaintext []byte) []byte
		Decrypt(ciphertext []byte) []byte
		Key() []byte
	}

	AesCTRCipherImpl struct {
		key   []byte
		block cipher.Block
	}
)

func NewAesCTRCipher(key string) AesCTRCipher {
	bytes := []byte(key)
	HEXKey := hex.EncodeToString(bytes)
	// The key MUST have 32 chars long
	if length := len(HEXKey); length < 32 {
		panic(fmt.Sprintf("cipher's key too short: %v", length))
	} else {
		HEXKey = HEXKey[:32]
	}
	bytes, err := hex.DecodeString(HEXKey)
	if err != nil {
		panic("unable to decode HEX key: " + err.Error())
	}
	block, err := aes.NewCipher(bytes)
	if err != nil {
		panic("unable to create cipher from key: " + err.Error())
	}
	return AesCTRCipherImpl{
		key:   bytes,
		block: block,
	}
}

func (aesCtrCipher AesCTRCipherImpl) Encrypt(plaintext []byte) []byte {
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	iv := ciphertext[:aes.BlockSize]
	_, err := io.ReadFull(rand.Reader, iv)
	if err != nil {
		panic("unable to create IV's cihper: " + err.Error())
	}

	stream := cipher.NewCTR(aesCtrCipher.block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext
}

func (aesCtrCipher AesCTRCipherImpl) Decrypt(ciphertext []byte) []byte {
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCTR(aesCtrCipher.block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext
}

func (aesCtrCipher AesCTRCipherImpl) Key() []byte {
	return aesCtrCipher.key
}
