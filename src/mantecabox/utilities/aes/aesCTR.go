// Package aes is a reference implementacion from https://gist.github.com/manishtpatel/8222606.
// In this package we process errors with panic because this cipher must be always working properly for this program
package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
)

var (
	KeyStr = "6368616e676520746869732070617373" // FIXME this key should be obtained in an external way to the program
	Key    []byte
)

func init() {
	bytes, err := hex.DecodeString(KeyStr)
	Key = bytes
	processErr(err)
}

func Encrypt(plaintext []byte) []byte {
	block, err := aes.NewCipher(Key)
	processErr(err)

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	iv := ciphertext[:aes.BlockSize]
	_, err = io.ReadFull(rand.Reader, iv)
	processErr(err)

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext
}

func Decrypt(ciphertext []byte) []byte {
	block, err := aes.NewCipher(Key)
	processErr(err)

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext
}

func processErr(err error) {
	if err != nil {
		panic(err)
	}
}
