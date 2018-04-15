package services

import (
	"database/sql"
	"encoding/base64"
	"os"
	"testing"

	"github.com/paveltrufi/mantecabox/dao/postgres"
	"github.com/paveltrufi/mantecabox/models"
	"github.com/paveltrufi/mantecabox/utilities"
	"github.com/paveltrufi/mantecabox/utilities/aes"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestMain(m *testing.M) {
	utilities.StartDockerPostgresDb()
	// os.Setenv("MANTECABOX_CONFIG_FILE", "configuration.test.json")
	// it must be initialized from the run script configuration

	code := m.Run()

	db := postgres.GetPgDb()
	cleanDb(db)
	os.Exit(code)
}

func TestRegisterUser(t *testing.T) {
	db := getDb(t)
	defer db.Close()
	testCases := []struct {
		name string
		test func(*testing.T)
	}{
		{
			"When the user to register has normal credentials, register it",
			func(t *testing.T) {
				expectedCredentials := models.Credentials{
					Username: "testuser",
					// base64(sha512(password))
					Password: "MTg2NEU3NTRCN0QyOENDOTk0OURDQkI1MEVFM0FFNEY3NTdCRjc1MTAwRjBDMkMzRTM3RDUxQ0Y0QURDNEVDREU0NDhCODQ2ODdEQTg3QjY5RTJGNkRCNTQwRUVFODMwNDM1MjY0RDlGNDcwNzc5MTQ4MUYyNUQ0NUUyOEQ5MTA=",
				}

				actualUser, err := RegisterUser(&expectedCredentials)
				require.NoError(t, err)
				require.Equal(t, expectedCredentials.Username, actualUser.Username)

				decodedExpectedPassword, err := base64.URLEncoding.DecodeString(expectedCredentials.Password)
				require.NoError(t, err)
				decodedActualPassword, err := base64.URLEncoding.DecodeString(actualUser.Password)
				require.NoError(t, err)
				err = bcrypt.CompareHashAndPassword(aes.Decrypt(decodedActualPassword), decodedExpectedPassword)
				require.NoError(t, err)
			},
		},
		{
			"When the user to register has no username, throw a bad username error",
			func(t *testing.T) {
				actualUser, err := RegisterUser(&models.Credentials{})
				require.Error(t, err)
				require.Equal(t, InvalidUsernameError, err.Error())
				require.Equal(t, models.User{}, actualUser)
			},
		},
		{
			"When the user to register has no password, throw a base64 error",
			func(t *testing.T) {
				actualUser, err := RegisterUser(&models.Credentials{Username: "testuser"})
				require.Error(t, err)
				require.Equal(t, models.User{}, actualUser)
			},
		},
		{
			"When the user to register has a non-hashed password, throw an invalid password error",
			func(t *testing.T) {
				actualUser, err := RegisterUser(&models.Credentials{
					Username: "testuser",
					// base64(password)
					Password: "bWFudGVjYWJveA==",
				})
				require.Error(t, err)
				require.Equal(t, InvalidPasswordError, err.Error())
				require.Equal(t, models.User{}, actualUser)
			},
		},
		{
			"When the user has a hashed password, but the algorithm used was not SHA-512, throw an invalid password error",
			func(t *testing.T) {
				actualUser, err := RegisterUser(&models.Credentials{
					Username: "testuser",
					// base64(sha256(password))
					Password: "MzFkYzhlYmMzZDhhN2U0ZjlhMzU4N2RkYWJkOGMxYmEwYjE5Yjc5ZjU2MWU1Yzk2MDhjYjQ4ZDRiMTRlOWFmMA==",
				})
				require.Error(t, err)
				require.Equal(t, InvalidPasswordError, err.Error())
				require.Equal(t, models.User{}, actualUser)
			},
		},
	}
	for _, testCase := range testCases {
		cleanDb(db)
		t.Run(testCase.name, testCase.test)
	}
}

func getDb(t *testing.T) *sql.DB {
	// Test preparation
	db := postgres.GetPgDb()
	require.NotNil(t, db)
	return db
}

func cleanDb(db *sql.DB) {
	db.Exec("DELETE FROM users")
}
