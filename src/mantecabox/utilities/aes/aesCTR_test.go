package aes

import (
	"io/ioutil"
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

func TestFileNewCTR(t *testing.T) {
	testFile, err := ioutil.ReadFile("/Users/raul/Go/mantecabox/files/raul_pairo@icloud.com/Calendario.pdf")
	require.NoError(t, err)
	encrypted := Encrypt(testFile)
	decrypted := Decrypt(encrypted)

	// To check the string is in base64, we try to decode it
	require.Equal(t, testFile, decrypted)
}
