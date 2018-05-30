package services

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"mantecabox/models"

	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v3"
)

var (
	successfulAttempt = models.LoginAttempt{
		User:       models.User{Credentials: models.Credentials{Email: testUserEmail, Password: correctPassword}},
		UserAgent:  null.String{NullString: sql.NullString{String: "Mozilla/5.0 (Linux; Android 6.0; Nexus 5X Build/MDB08L) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.76 Mobile Safari/537.36", Valid: true}},
		IP:         null.String{NullString: sql.NullString{String: "127.0.0.1", Valid: true}},
		Successful: true,
	}
	unsuccessfulAttempt = models.LoginAttempt{
		User:       successfulAttempt.User,
		UserAgent:  successfulAttempt.UserAgent,
		IP:         successfulAttempt.IP,
		Successful: false,
	}
)

func TestProcessLoginAttempt(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "When user has three unsuccessful login attempts, at the fourth one return an error",
			test: func(t *testing.T) {
				userDao.Create(&unsuccessfulAttempt.User)

				for i := 0; i < 3; i++ {
					err := ProcessLoginAttempt(&successfulAttempt)
					require.NoError(t, err)
				}
				err := ProcessLoginAttempt(&unsuccessfulAttempt)
				require.NoError(t, err)
				err = ProcessLoginAttempt(&unsuccessfulAttempt)
				require.NoError(t, err)
				err = ProcessLoginAttempt(&unsuccessfulAttempt)
				require.Error(t, err)
				err = ProcessLoginAttempt(&unsuccessfulAttempt)
				require.Equal(t, TooManyAttemptsErr, err)
			},
		},
		{
			name: "When user has three unsuccessful login attempts, and then one successful, return an error",
			test: func(t *testing.T) {
				userDao.Create(&unsuccessfulAttempt.User)

				for i := 0; i < 3; i++ {
					err := ProcessLoginAttempt(&successfulAttempt)
					require.NoError(t, err)
				}
				err := ProcessLoginAttempt(&unsuccessfulAttempt)
				require.NoError(t, err)
				err = ProcessLoginAttempt(&unsuccessfulAttempt)
				require.NoError(t, err)
				err = ProcessLoginAttempt(&unsuccessfulAttempt)
				require.Error(t, err)
				err = ProcessLoginAttempt(&successfulAttempt)
				require.Equal(t, errors.New(fmt.Sprintf("Login for user %v blocked for the next %.2f minutes", successfulAttempt.User.Email, timeLimit.Minutes())), err)
			},
		},
		{
			name: "When user has three successful login attempts, don't return any error at any of them",
			test: func(t *testing.T) {
				userDao.Create(&successfulAttempt.User)

				for i := 0; i < 3; i++ {
					err := ProcessLoginAttempt(&successfulAttempt)
					require.NoError(t, err)
				}
			},
		},
		// Missing some cases to check if some report function was called,
		// but for that I would need to mock them and I don't have enough time
	}
	for _, testCase := range tests {
		db := getDb(t)
		cleanDb(db)
		t.Run(testCase.name, testCase.test)
	}
}

func Test_sendNewRegisteredDeviceActivity(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "When sending a new device report, no error must be returned",
			test: func(t *testing.T) {
				userDao.Create(&successfulAttempt.User)
				err := sendNewRegisteredDeviceActivity(&successfulAttempt)
				require.NoError(t, err)
			},
		},
		{
			name: "When sending a new device report without User-Agent, no error must be returned",
			test: func(t *testing.T) {
				attemptWithoutUserAgent := successfulAttempt
				attemptWithoutUserAgent.UserAgent = null.String{}
				userDao.Create(&attemptWithoutUserAgent.User)
				err := sendNewRegisteredDeviceActivity(&attemptWithoutUserAgent)
				require.NoError(t, err)
			},
		},
		{
			name: "When sending a new device report without IPv4, no error must be returned",
			test: func(t *testing.T) {
				attemptWithoutUserAgent := successfulAttempt
				attemptWithoutUserAgent.IP = null.String{}
				userDao.Create(&attemptWithoutUserAgent.User)
				err := sendNewRegisteredDeviceActivity(&attemptWithoutUserAgent)
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range tests {
		db := getDb(t)
		cleanDb(db)
		t.Run(testCase.name, testCase.test)
	}
}

func Test_sendSuspiciousActivityReport(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "When sending a suspicious activity report, no error must be returned",
			test: func(t *testing.T) {
				userDao.Create(&successfulAttempt.User)
				err := sendSuspiciousActivityReport(&unsuccessfulAttempt)
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range tests {
		db := getDb(t)
		cleanDb(db)
		t.Run(testCase.name, testCase.test)
	}
}
