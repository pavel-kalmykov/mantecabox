package postgres

import (
	"database/sql"
	"testing"

	"mantecabox/models"

	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v3"
)

const (
	testuserWithFourAttemptsInsert = testUserInsert + `insert into login_attempts ("user", user_agent, ip, successful) values 
  ('testuser1', 'user-agent1', '127.0.0.1', true),
  ('testuser1', 'user-agent2', '127.0.0.1', true),
  ('testuser1', 'user-agent1', '192.168.0.160', false),
  ('testuser1', null, null, false);`
)

var (
	dao               = LoginAttemptPgDao{}
	loginAttemptTest1 = models.LoginAttempt{
		User:       models.User{Credentials: models.Credentials{Email: "testuser1", Password: "testpassword1"}},
		UserAgent:  null.String{NullString: sql.NullString{String: "user-agent1", Valid: true}},
		IP:         null.String{NullString: sql.NullString{String: "127.0.0.1", Valid: true}},
		Successful: true,
	}
	loginAttemptTest2 = models.LoginAttempt{
		User:       models.User{Credentials: models.Credentials{Email: "testuser1", Password: "testpassword1"}},
		UserAgent:  null.String{NullString: sql.NullString{String: "user-agent2", Valid: true}},
		IP:         null.String{NullString: sql.NullString{String: "127.0.0.1", Valid: true}},
		Successful: true,
	}
	loginAttemptTest3 = models.LoginAttempt{
		User:       models.User{Credentials: models.Credentials{Email: "testuser1", Password: "testpassword1"}},
		UserAgent:  null.String{NullString: sql.NullString{String: "user-agent1", Valid: true}},
		IP:         null.String{NullString: sql.NullString{String: "192.168.0.160", Valid: true}},
		Successful: false,
	}
	loginAttemptTest4 = models.LoginAttempt{
		User:       models.User{Credentials: models.Credentials{Email: "testuser1", Password: "testpassword1"}},
		UserAgent:  null.String{NullString: sql.NullString{Valid: false}},
		IP:         null.String{NullString: sql.NullString{Valid: false}},
		Successful: false,
	}
)

func TestLoginAttemptPgDao_GetByUser(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name      string
		prepQuery string
		args      args
		want      []models.LoginAttempt
	}{
		{
			name:      "When a user has four login attempts, return all of them",
			prepQuery: testuserWithFourAttemptsInsert,
			args:      args{email: "testuser1"},
			want: []models.LoginAttempt{
				loginAttemptTest1,
				loginAttemptTest2,
				loginAttemptTest3,
				loginAttemptTest4,
			},
		},
		{
			name:      "When there are no login attempts for a user, return an empty array",
			prepQuery: "",
			args:      args{email: "testuser1"},
			want:      []models.LoginAttempt{},
		},
		{
			name:      "When you look for the login attempts of an unexistent user, return an empty array",
			prepQuery: testuserWithFourAttemptsInsert,
			args:      args{email: "unexistent"},
			want:      []models.LoginAttempt{},
		},
	}
	db := getDb(t)
	defer db.Close()
	for _, testCase := range tests {
		cleanAndPopulateDb(db, testCase.prepQuery, t)
		t.Run(testCase.name, func(t *testing.T) {
			got, err := dao.GetByUser(testCase.args.email)
			require.NoError(t, err)
			require.Equal(t, len(testCase.want), len(got))
			for i, expected := range testCase.want {
				requireUserAttemptEqual(t, expected, got[i])
			}
		})
	}
}

func TestLoginAttemptPgDao_GetLastNByUser(t *testing.T) {
	type args struct {
		email string
		n     int
	}
	tests := []struct {
		name      string
		prepQuery string
		args      args
		want      []models.LoginAttempt
	}{
		{
			name:      "When a user has four login attempts and you want the last three, return the last three of them",
			prepQuery: testuserWithFourAttemptsInsert,
			args:      args{email: "testuser1", n: 3},
			want: []models.LoginAttempt{
				loginAttemptTest2,
				loginAttemptTest3,
				loginAttemptTest4,
			},
		},
		{
			name:      "When a user has four login attempts and you want the last zero, return an empty array",
			prepQuery: testuserWithFourAttemptsInsert,
			args:      args{email: "testuser1", n: 0},
			want:      []models.LoginAttempt{},
		},
		{
			name:      "When a user has four login attempts and you want less than zero, return all of them",
			prepQuery: testuserWithFourAttemptsInsert,
			args:      args{email: "testuser1", n: -1},
			want: []models.LoginAttempt{
				loginAttemptTest1,
				loginAttemptTest2,
				loginAttemptTest3,
				loginAttemptTest4,
			},
		},
	}
	db := getDb(t)
	defer db.Close()
	for _, testCase := range tests {
		cleanAndPopulateDb(db, testCase.prepQuery, t)
		t.Run(testCase.name, func(t *testing.T) {
			got, err := dao.GetLastNByUser(testCase.args.email, testCase.args.n)
			require.NoError(t, err)
			require.Equal(t, len(testCase.want), len(got))
			for i, expected := range testCase.want {
				requireUserAttemptEqual(t, expected, got[i])
			}
		})
	}
}

func TestLoginAttemptPgDao_Create(t *testing.T) {
	successfulAttempt := models.LoginAttempt{
		User:       models.User{Credentials: models.Credentials{Email: "testuser1", Password: "testpassword1"}},
		UserAgent:  null.String{NullString: sql.NullString{String: "Mozilla/5.0 (Linux; Android 6.0; Nexus 5X Build/MDB08L) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.76 Mobile Safari/537.36", Valid: true}},
		IP:         null.String{NullString: sql.NullString{String: "127.0.0.1", Valid: true}},
		Successful: true,
	}
	unsuccessfulAttempt := successfulAttempt
	unsuccessfulAttempt.Successful = false
	attemptWithoutIps := successfulAttempt
	attemptWithoutIps.IP = null.String{}
	attemptWithoutUserAgent := successfulAttempt
	attemptWithoutUserAgent.UserAgent = null.String{}
	attemptWithMalformedIp := successfulAttempt
	attemptWithMalformedIp.IP = null.String{NullString: sql.NullString{String: "localhostlocalhostlocalhostlocalhostlocalhostlocalhost", Valid: true}}

	type args struct {
		attempt *models.LoginAttempt
	}
	tests := []struct {
		name      string
		prepQuery string
		args      args
		want      *models.LoginAttempt
		wantErr   bool
	}{
		{
			name:      "Create a normal attempt with every field",
			prepQuery: testUserInsert,
			args:      args{attempt: &successfulAttempt},
			want:      &successfulAttempt,
			wantErr:   false,
		},
		{
			name:      "Create an unsuccessful attempt with every field",
			prepQuery: testUserInsert,
			args:      args{attempt: &unsuccessfulAttempt},
			want:      &unsuccessfulAttempt,
			wantErr:   false,
		},
		{
			name:      "Create an attempt without IPs",
			prepQuery: testUserInsert,
			args:      args{attempt: &attemptWithoutIps},
			want:      &attemptWithoutIps,
			wantErr:   false,
		},
		{
			name:      "Create an attempt without User-Agent",
			prepQuery: testUserInsert,
			args:      args{attempt: &attemptWithoutUserAgent},
			want:      &attemptWithoutUserAgent,
			wantErr:   false,
		},
		{
			name:      "Create an attempt with malformed IP",
			prepQuery: testUserInsert,
			args:      args{attempt: &attemptWithMalformedIp},
			want:      &attemptWithMalformedIp,
			wantErr:   true,
		},
	}
	db := getDb(t)
	defer db.Close()
	for _, testCase := range tests {
		cleanAndPopulateDb(db, testCase.prepQuery, t)
		t.Run(testCase.name, func(t *testing.T) {
			got, err := dao.Create(testCase.args.attempt)
			if testCase.wantErr {
				require.Error(t, err)
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				requireUserAttemptEqual(t, *testCase.want, got)
			}
		})
	}
}

func TestLoginAttemptPgDao_GetSimilarAttempts(t *testing.T) {
	attemptFromNewPlace := loginAttemptTest1
	attemptFromNewPlace.IP.SetValid("209.173.53.167")

	type args struct {
		attempt *models.LoginAttempt
	}
	tests := []struct {
		name      string
		prepQuery string
		args      args
		want      []models.LoginAttempt
		wantErr   bool
	}{
		{
			name:      "When a new login attempt has the same user-agent, IPv4 and IPv6 than an existing one, return it",
			prepQuery: testuserWithFourAttemptsInsert,
			args:      args{attempt: &loginAttemptTest1},
			want:      []models.LoginAttempt{loginAttemptTest1},
			wantErr:   false,
		},
		{
			name:      "When a new login attempt is from a new place, return an empty array",
			prepQuery: testuserWithFourAttemptsInsert,
			args:      args{attempt: &attemptFromNewPlace},
			want:      []models.LoginAttempt{},
			wantErr:   false,
		},
	}
	db := getDb(t)
	defer db.Close()
	for _, testCase := range tests {
		cleanAndPopulateDb(db, testCase.prepQuery, t)
		t.Run(testCase.name, func(t *testing.T) {
			got, err := dao.GetSimilarAttempts(testCase.args.attempt)
			if testCase.wantErr {
				require.Error(t, err)
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(testCase.want), len(got))
				for i, expected := range testCase.want {
					requireUserAttemptEqual(t, expected, got[i])
				}
			}
		})
	}
}

func requireUserAttemptEqual(t *testing.T, expected models.LoginAttempt, actual models.LoginAttempt) {
	require.Equal(t, expected.User.Credentials, actual.User.Credentials)
	require.Equal(t, expected.UserAgent, actual.UserAgent)
	require.Equal(t, expected.IP, actual.IP)
	require.Equal(t, expected.Successful, actual.Successful)
}
