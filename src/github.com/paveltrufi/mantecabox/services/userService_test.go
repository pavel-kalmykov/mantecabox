package services

import (
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/paveltrufi/mantecabox/dao/postgres"
	"github.com/paveltrufi/mantecabox/models"
	"github.com/paveltrufi/mantecabox/utilities"
	"github.com/paveltrufi/mantecabox/utilities/aes"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

const (
	testUserEmail    = "hello@example.com"
	testUser2Email   = "hello2@example.com"
	updatedUserEmail = "updated@example.com"
)

var correctPassword = "testsecret"

func init() {
	sum512 := sha512.Sum512([]byte(correctPassword))
	str := strings.ToUpper(hex.EncodeToString(sum512[:]))
	correctPassword = base64.URLEncoding.EncodeToString([]byte(str))
}

func TestMain(m *testing.M) {
	utilities.StartDockerPostgresDb()

	code := m.Run()

	db := postgres.GetPgDb()
	cleanDb(db)
	os.Exit(code)
}

func TestGetUsers(t *testing.T) {
	db := getDb(t)
	defer db.Close()
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			"Get users test",
			func(t *testing.T) {
				actualUsers, err := GetUsers()
				require.NoError(t, err)
				require.NotEmpty(t, actualUsers)
			},
		},
	}
	for _, tt := range tests {
		cleanDb(db)
		userDao.Create(&models.User{Credentials: models.Credentials{Email: testUserEmail, Password: "testpassword"}})
		userDao.Create(&models.User{Credentials: models.Credentials{Email: testUser2Email, Password: "testpassword2"}})
		t.Run(tt.name, tt.test)
	}

}

func TestGetUser(t *testing.T) {
	db := getDb(t)
	defer db.Close()
	expectedCredentials := models.Credentials{
		Email:    testUserEmail,
		Password: correctPassword,
	}
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			"Get user test",
			func(t *testing.T) {
				actualUser, err := GetUser(testUserEmail)
				require.NoError(t, err)
				require.Equal(t, testUserEmail, actualUser.Email)
				decodedExpectedPassword, err := base64.URLEncoding.DecodeString(expectedCredentials.Password)
				require.NoError(t, err)
				decodedActualPassword, err := base64.URLEncoding.DecodeString(actualUser.Password)
				require.NoError(t, err)
				err = bcrypt.CompareHashAndPassword(aes.Decrypt(decodedActualPassword), decodedExpectedPassword)
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		cleanDb(db)
		RegisterUser(&expectedCredentials)
		t.Run(tt.name, tt.test)
	}
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
					Email:    testUserEmail,
					Password: correctPassword,
				}

				actualUser, err := RegisterUser(&expectedCredentials)
				require.NoError(t, err)
				require.Equal(t, expectedCredentials.Email, actualUser.Email)

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
				require.True(t, strings.HasPrefix(err.Error(), InvalidEmailError.Error()))
				require.Equal(t, models.User{}, actualUser)
			},
		},
		{
			"When the user to register has no password, throw a base64 error",
			func(t *testing.T) {
				actualUser, err := RegisterUser(&models.Credentials{Email: testUserEmail})
				require.Error(t, err)
				require.Equal(t, models.User{}, actualUser)
			},
		},
		{
			"When the user to register has a non-hashed password, throw an invalid password error",
			func(t *testing.T) {
				actualUser, err := RegisterUser(&models.Credentials{
					Email: testUserEmail,
					// base64(password)
					Password: "bWFudGVjYWJveA==",
				})
				require.Error(t, err)
				require.Equal(t, InvalidPasswordError, err)
				require.Equal(t, models.User{}, actualUser)
			},
		},
		{
			"When the user has a hashed password, but the algorithm used was not SHA-512, throw an invalid password error",
			func(t *testing.T) {
				actualUser, err := RegisterUser(&models.Credentials{
					Email: testUserEmail,
					// base64(sha256(password))
					Password: "MzFkYzhlYmMzZDhhN2U0ZjlhMzU4N2RkYWJkOGMxYmEwYjE5Yjc5ZjU2MWU1Yzk2MDhjYjQ4ZDRiMTRlOWFmMA==",
				})
				require.Error(t, err)
				require.Equal(t, InvalidPasswordError, err)
				require.Equal(t, models.User{}, actualUser)
			},
		},
	}
	for _, testCase := range testCases {
		cleanDb(db)
		t.Run(testCase.name, testCase.test)
	}
}

func TestModifyUser(t *testing.T) {
	db := getDb(t)
	defer db.Close()
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			"When the user to modify exists and has normal credentials, modify it",
			func(t *testing.T) {
				expectedUser := models.User{
					Credentials: models.Credentials{
						Email:    updatedUserEmail,
						Password: correctPassword,
					},
				}

				actualUser, err := ModifyUser(testUserEmail, &expectedUser)
				require.NoError(t, err)
				require.Equal(t, expectedUser.Email, actualUser.Email)

				decodedExpectedPassword, err := base64.URLEncoding.DecodeString(expectedUser.Password)
				require.NoError(t, err)
				decodedActualPassword, err := base64.URLEncoding.DecodeString(actualUser.Password)
				require.NoError(t, err)
				err = bcrypt.CompareHashAndPassword(aes.Decrypt(decodedActualPassword), decodedExpectedPassword)
				require.NoError(t, err)
			},
		},
		{
			"When the user to modify exists and has no username, throw a bad username error",
			func(t *testing.T) {
				actualUser, err := ModifyUser(testUserEmail, &models.User{})
				require.Error(t, err)
				require.True(t, strings.HasPrefix(err.Error(), InvalidEmailError.Error()))
				require.Equal(t, models.User{}, actualUser)
			},
		},
		{
			"When the user to modify exists and has no password, throw a base64 error",
			func(t *testing.T) {
				actualUser, err := ModifyUser(testUserEmail, &models.User{
					Credentials: models.Credentials{
						Email: testUserEmail,
					},
				})
				require.Error(t, err)
				require.Equal(t, models.User{}, actualUser)
			},
		},
		{
			"When the user to modify exists and has a non-hashed password, throw an invalid password error",
			func(t *testing.T) {
				actualUser, err := ModifyUser(testUserEmail, &models.User{
					Credentials: models.Credentials{
						Email: testUserEmail,
						// base64(password)
						Password: "bWFudGVjYWJveA==",
					},
				})
				require.Error(t, err)
				require.Equal(t, InvalidPasswordError, err)
				require.Equal(t, models.User{}, actualUser)
			},
		},
		{
			"When the user has a hashed password, but the algorithm used was not SHA-512, throw an invalid password error",
			func(t *testing.T) {
				actualUser, err := RegisterUser(&models.Credentials{
					Email: testUserEmail,
					// base64(sha256(password))
					Password: "MzFkYzhlYmMzZDhhN2U0ZjlhMzU4N2RkYWJkOGMxYmEwYjE5Yjc5ZjU2MWU1Yzk2MDhjYjQ4ZDRiMTRlOWFmMA==",
				})
				require.Error(t, err)
				require.Equal(t, InvalidPasswordError, err)
				require.Equal(t, models.User{}, actualUser)
			},
		},
		{
			"When the user does not exist, throw an error",
			func(t *testing.T) {
				expectedUser := models.User{
					Credentials: models.Credentials{
						Email:    updatedUserEmail,
						Password: correctPassword,
					},
				}

				actualUser, err := ModifyUser("nonexistentuser", &expectedUser)
				require.Error(t, err)
				require.Equal(t, models.User{}, actualUser)
			},
		},
	}
	for _, tt := range tests {
		cleanDb(db)
		userDao.Create(&models.User{Credentials: models.Credentials{Email: testUserEmail, Password: "testpassword"}})
		t.Run(tt.name, tt.test)
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

func TestDeleteUser(t *testing.T) {
	db := getDb(t)
	defer db.Close()
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			"When the user exists, delete it",
			func(t *testing.T) {
				err := DeleteUser(testUserEmail)
				require.NoError(t, err)
			},
		},
		{
			name: "When the user doesn't exist, return an error",
			test: func(t *testing.T) {
				err := DeleteUser("nonexistent")
				require.Equal(t, sql.ErrNoRows, err)
			},
		},
	}
	for _, tt := range tests {
		cleanDb(db)
		userDao.Create(&models.User{Credentials: models.Credentials{Email: testUserEmail, Password: "testpassword"}})
		t.Run(tt.name, tt.test)
	}
}

func TestUserExists(t *testing.T) {
	db := getDb(t)
	defer db.Close()
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "When the user exists and password is correct, return true",
			test: func(t *testing.T) {
				user, exists := UserExists(testUserEmail, correctPassword)
				require.Equal(t, testUserEmail, user.Email)
				require.True(t, exists)
			},
		},
		{
			name: "When the user exists but password is not in base64, return false",
			test: func(t *testing.T) {
				user, exists := UserExists(testUserEmail, "testpassword")
				require.Equal(t, testUserEmail, user.Email)
				require.False(t, exists)
			},
		},
		{
			name: "When the user exists but password is incorrect, return false",
			test: func(t *testing.T) {
				user, exists := UserExists(testUserEmail, "MzFkYzhlYmMzZDhhN2U0ZjlhMzU4N2RkYWJkOGMxYmEwYjE5Yjc5ZjU2MWU1Yzk2MDhjYjQ4ZDRiMTRlOWFmMA==")
				require.Equal(t, testUserEmail, user.Email)
				require.False(t, exists)
			},
		},
		{
			name: "When the user doesn't exist, return false",
			test: func(t *testing.T) {
				user, exists := UserExists("nonexistent", "")
				require.Equal(t, "nonexistent", user.Email)
				require.False(t, exists)
			},
		},
	}
	for _, tt := range tests {
		cleanDb(db)
		RegisterUser(&models.Credentials{
			Email:    testUserEmail,
			Password: correctPassword,
		})
		t.Run(tt.name, tt.test)
	}
}

func TestTwoFactorMatchesAndIsNotOutdated(t *testing.T) {
	type args struct {
		code1  string
		code2  string
		expire time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "When codes match and is not expired, return true",
			args: args{code1: "123456", code2: "123456", expire: time.Now().Add(-time.Minute * 2)},
			want: true,
		},
		{
			name: "When codes match and is expired, return false",
			args: args{code1: "123456", code2: "123456", expire: time.Now().Add(-time.Minute * 6)},
			want: false,
		},
		{
			name: "When codes don't match and is not expired, return false",
			args: args{code1: "123456", code2: "123457", expire: time.Now().Add(-time.Minute * 2)},
			want: false,
		},
		{
			name: "When codes don't match and is expired, return false",
			args: args{code1: "123456", code2: "123457", expire: time.Now().Add(-time.Minute * 6)},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, TwoFactorMatchesAndIsNotOutdated(tt.args.code1, tt.args.code2, tt.args.expire))
		})
	}
}

func TestGenerate2FACodeAndSaveToUser(t *testing.T) {
	db := getDb(t)
	defer db.Close()
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "When generating a 2FA code to an existent user, return the user with the code and its timestamp",
			test: func(t *testing.T) {
				createdUser, err := userDao.Create(&models.User{
					Credentials: models.Credentials{Email: testUserEmail, Password: correctPassword},
				})
				require.NoError(t, err)
				userWithCode, err := Generate2FACodeAndSaveToUser(&createdUser)
				require.NoError(t, err)
				twoFactorAuth := userWithCode.TwoFactorAuth.ValueOrZero()
				twoFactorTime := userWithCode.TwoFactorTime.ValueOrZero()
				require.NotEmpty(t, twoFactorAuth)
				require.NotEmpty(t, twoFactorTime)
				require.True(t, len(twoFactorAuth) == 6)
				require.True(t, twoFactorTime.Sub(time.Now()) < time.Minute)
			},
		},
		{
			name: "When generating a 2FA code to an unexistent user, return an error",
			test: func(t *testing.T) {
				userWithCode, err := Generate2FACodeAndSaveToUser(&models.User{
					Credentials: models.Credentials{Email: testUserEmail, Password: correctPassword},
				})
				require.Error(t, err)
				require.Empty(t, userWithCode)
			},
		},
	}
	for _, tt := range tests {
		cleanDb(db)
		t.Run(tt.name, tt.test)
	}
}

func TestSend2FAEmail(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "When the 2FA is sent correctly, return no error",
			test: func(t *testing.T) {
				require.NoError(t, Send2FAEmail("mantecabox@gmail.com", "123456"))
			},
		},
		{
			name: "When the email does not exist, return an error",
			test: func(t *testing.T) {
				require.Error(t, Send2FAEmail("unexistent@error.ko", "123456"))
			},
		},
		{
			name: "When the code is empty, return an error",
			test: func(t *testing.T) {
				require.Equal(t, Empty2FACodeError, Send2FAEmail("mantecabox@gmail.com", ""))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}
