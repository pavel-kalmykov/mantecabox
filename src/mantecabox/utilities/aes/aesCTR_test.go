package aes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExampleNewCTR(t *testing.T) {
	const testString = "mantecabox"

	encrypted := Encrypt([]byte(testString))
	decrypted := Decrypt(encrypted)

	// To check the string is in base64, we try to decode it
	require.Equal(t, testString, string(decrypted))
}
