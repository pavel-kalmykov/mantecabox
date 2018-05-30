package services

import (
	"database/sql"
	"testing"

	"mantecabox/models"

	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v3"
)

var (
	successfulAttempt = models.LoginAttempt{
		User:       models.User{Credentials: models.Credentials{Email: testUserEmail, Password: correctPassword}},
		UserAgent:  null.String{NullString: sql.NullString{String: "Mozilla/5.0 (Linux; Android 6.0; Nexus 5X Build/MDB08L) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.76 Mobile Safari/537.36", Valid: true}},
		IPv4:       null.String{NullString: sql.NullString{String: "127.0.0.1", Valid: true}},
		IPv6:       null.String{NullString: sql.NullString{String: "::1", Valid: true}},
		Successful: true,
	}
	unsuccessfulAttempt = models.LoginAttempt{
		User:       successfulAttempt.User,
		UserAgent:  successfulAttempt.UserAgent,
		IPv4:       successfulAttempt.IPv4,
		IPv6:       successfulAttempt.IPv6,
		Successful: false,
	}
)

func TestProcessLoginAttempt(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "When user has three unsuccessful login attempts, at the third one return an error",
			test: func(t *testing.T) {
				userDao.Create(&unsuccessfulAttempt.User)

				err := ProcessLoginAttempt(&unsuccessfulAttempt)
				require.NoError(t, err)
				err = ProcessLoginAttempt(&unsuccessfulAttempt)
				require.NoError(t, err)
				err = ProcessLoginAttempt(&unsuccessfulAttempt)
				require.Equal(t, TooManyAttemptsErr, err)
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
}

func Test_sendSuspiciousActivityReport(t *testing.T) {
}
