package utilities

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

var testAesCTRCipher = NewAesCTRCipher("this is an AES cipher in CTR mode key")

func TestExampleNewCTR(t *testing.T) {
	const testString = "mantecabox"

	encrypted := testAesCTRCipher.Encrypt([]byte(testString))
	decrypted := testAesCTRCipher.Decrypt(encrypted)

	// To check the string is in base64, we try to decode it
	require.Equal(t, testString, string(decrypted))
}

func TestFileNewCTR(t *testing.T) {
	testFile, err := ioutil.ReadAll(bytes.NewReader([]byte("Fichero inventado de Mantecabox")))
	require.NoError(t, err)
	encrypted := testAesCTRCipher.Encrypt(testFile)
	decrypted := testAesCTRCipher.Decrypt(encrypted)

	// To check the string is in base64, we try to decode it
	require.Equal(t, testFile, decrypted)
}

func TestNewAesCTRCipher(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want AesCTRCipher
	}{
		{
			name: "When a key is passed, return the cipher",
			args: args{key: "0123456789ABCDEF"},
			want: AesCTRCipherImpl{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.IsType(t, tt.want, NewAesCTRCipher(tt.args.key))
		})
	}
}
