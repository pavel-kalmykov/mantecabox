package postgres

import (
	"database/sql"
	"testing"

	"mantecabox/models"

	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v3"
)

const (
	testUserWithFourAttemptsInsert = `insert into users (email, password) values ('testuser', 'testpass');
insert into login_attempts ("user", ipv4, successful) values
  ('testuser', '216.3.128.12', true),
  ('testuser', '216.3.128.12', true),
  ('testuser', '216.3.128.12', false),
  ('testuser', '216.3.128.12', false);`
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
			prepQuery: testUserWithFourAttemptsInsert,
			args:      args{email: "testuser"},
			want: []models.LoginAttempt{
				{
					User:       models.User{Credentials: models.Credentials{Email: "testuser", Password: "testpass"}},
					IPv4:       null.String{NullString: sql.NullString{String: "216.3.128.12", Valid: true}},
					Successful: true,
				},
				{
					User:       models.User{Credentials: models.Credentials{Email: "testuser", Password: "testpass"}},
					IPv4:       null.String{NullString: sql.NullString{String: "216.3.128.12", Valid: true}},
					Successful: true,
				},
				{
					User:       models.User{Credentials: models.Credentials{Email: "testuser", Password: "testpass"}},
					IPv4:       null.String{NullString: sql.NullString{String: "216.3.128.12", Valid: true}},
					Successful: false,
				},
				{
					User:       models.User{Credentials: models.Credentials{Email: "testuser", Password: "testpass"}},
					IPv4:       null.String{NullString: sql.NullString{String: "216.3.128.12", Valid: true}},
					Successful: false,
				},
			},
		},
		{
			name:      "When there are no login attempts for a user, return an empty array",
			prepQuery: "",
			args:      args{email: "testuser"},
			want:      []models.LoginAttempt{},
		},
		{
			name:      "When you look for the login attempts of an unexistent user, return an empty array",
			prepQuery: testUserWithFourAttemptsInsert,
			args:      args{email: "unexistent"},
			want:      []models.LoginAttempt{},
		},
	}
	db := getDb(t)
	defer db.Close()
	for _, testCase := range tests {
		cleanAndPopulateDb(db, testCase.prepQuery, t)
		t.Run(testCase.name, func(t *testing.T) {
			dao := LoginAttemptPgDao{}
			got, err := dao.GetByUser(testCase.args.email)
			require.NoError(t, err)
			require.Equal(t, len(testCase.want), len(got))
			for i, expected := range testCase.want {
				require.Equal(t, expected.User.Credentials, got[i].User.Credentials)
				require.Equal(t, expected.IPv4, got[i].IPv4)
				require.Equal(t, expected.Successful, got[i].Successful)
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
			prepQuery: testUserWithFourAttemptsInsert,
			args:      args{email: "testuser", n: 3},
			want: []models.LoginAttempt{
				{
					User:       models.User{Credentials: models.Credentials{Email: "testuser", Password: "testpass"}},
					IPv4:       null.String{NullString: sql.NullString{String: "216.3.128.12", Valid: true}},
					Successful: true,
				},
				{
					User:       models.User{Credentials: models.Credentials{Email: "testuser", Password: "testpass"}},
					IPv4:       null.String{NullString: sql.NullString{String: "216.3.128.12", Valid: true}},
					Successful: false,
				},
				{
					User:       models.User{Credentials: models.Credentials{Email: "testuser", Password: "testpass"}},
					IPv4:       null.String{NullString: sql.NullString{String: "216.3.128.12", Valid: true}},
					Successful: false,
				},
			},
		},
		{
			name:      "When a user has four login attempts and you want the last zero, return an empty array",
			prepQuery: testUserWithFourAttemptsInsert,
			args:      args{email: "testuser", n: 3},
			want: []models.LoginAttempt{
				{
					User:       models.User{Credentials: models.Credentials{Email: "testuser", Password: "testpass"}},
					IPv4:       null.String{NullString: sql.NullString{String: "216.3.128.12", Valid: true}},
					Successful: true,
				},
				{
					User:       models.User{Credentials: models.Credentials{Email: "testuser", Password: "testpass"}},
					IPv4:       null.String{NullString: sql.NullString{String: "216.3.128.12", Valid: true}},
					Successful: false,
				},
				{
					User:       models.User{Credentials: models.Credentials{Email: "testuser", Password: "testpass"}},
					IPv4:       null.String{NullString: sql.NullString{String: "216.3.128.12", Valid: true}},
					Successful: false,
				},
			},
		},
		{
			name:      "When a user has four login attempts and you want less than zero, return all of them",
			prepQuery: testUserWithFourAttemptsInsert,
			args:      args{email: "testuser", n: -1},
			want: []models.LoginAttempt{
				{
					User:       models.User{Credentials: models.Credentials{Email: "testuser", Password: "testpass"}},
					IPv4:       null.String{NullString: sql.NullString{String: "216.3.128.12", Valid: true}},
					Successful: true,
				},
				{
					User:       models.User{Credentials: models.Credentials{Email: "testuser", Password: "testpass"}},
					IPv4:       null.String{NullString: sql.NullString{String: "216.3.128.12", Valid: true}},
					Successful: true,
				},
				{
					User:       models.User{Credentials: models.Credentials{Email: "testuser", Password: "testpass"}},
					IPv4:       null.String{NullString: sql.NullString{String: "216.3.128.12", Valid: true}},
					Successful: false,
				},
				{
					User:       models.User{Credentials: models.Credentials{Email: "testuser", Password: "testpass"}},
					IPv4:       null.String{NullString: sql.NullString{String: "216.3.128.12", Valid: true}},
					Successful: false,
				},
			},
		},
	}
	db := getDb(t)
	defer db.Close()
	for _, testCase := range tests {
		cleanAndPopulateDb(db, testCase.prepQuery, t)
		t.Run(testCase.name, func(t *testing.T) {
			dao := LoginAttemptPgDao{}
			got, err := dao.GetLastNByUser(testCase.args.email, testCase.args.n)
			require.NoError(t, err)
			require.Equal(t, len(testCase.want), len(got))
			for i, expected := range testCase.want {
				require.Equal(t, expected.User.Credentials, got[i].User.Credentials)
				require.Equal(t, expected.IPv4, got[i].IPv4)
				require.Equal(t, expected.Successful, got[i].Successful)
			}
		})
	}
}

func TestLoginAttemptPgDao_Create(t *testing.T) {
	successfulAttempt := models.LoginAttempt{
		User:       models.User{Credentials: models.Credentials{Email: "testuser1", Password: "testpassword1"}},
		UserAgent:  null.String{NullString: sql.NullString{String: "Mozilla/5.0 (Linux; Android 6.0; Nexus 5X Build/MDB08L) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.76 Mobile Safari/537.36", Valid: true}},
		IPv4:       null.String{NullString: sql.NullString{String: "127.0.0.1", Valid: true}},
		IPv6:       null.String{NullString: sql.NullString{String: "0000:0000:0000:0000:0000:0000", Valid: true}},
		Successful: true,
	}
	unsuccessfulAttempt := successfulAttempt
	unsuccessfulAttempt.Successful = false
	attemptWithoutIps := successfulAttempt
	attemptWithoutIps.IPv4 = null.String{}
	attemptWithoutIps.IPv6 = null.String{}
	attemptWithoutUserAgent := successfulAttempt
	attemptWithoutUserAgent.UserAgent = null.String{}
	attemptWithMalformedIpv4 := successfulAttempt
	attemptWithMalformedIpv4.IPv4 = null.String{NullString: sql.NullString{String: "localhost 127.0.0.1", Valid: true}}
	attemptWithMalformedIpv6 := successfulAttempt
	attemptWithMalformedIpv6.IPv6 = null.String{NullString: sql.NullString{String: "127.0.0.1. 0000:0000:0000:0000:0000:0000", Valid: true}}

	type args struct {
		attempt models.LoginAttempt
	}
	tests := []struct {
		name      string
		prepQuery string
		args      args
		want      models.LoginAttempt
		wantErr   bool
	}{
		{
			name:      "Create a normal attempt with every field",
			prepQuery: testUserInsert,
			args:      args{attempt: successfulAttempt},
			want:      successfulAttempt,
			wantErr:   false,
		},
		{
			name:      "Create an unsuccessful attempt with every field",
			prepQuery: testUserInsert,
			args:      args{attempt: unsuccessfulAttempt},
			want:      unsuccessfulAttempt,
			wantErr:   false,
		},
		{
			name:      "Create an attempt without IPs",
			prepQuery: testUserInsert,
			args:      args{attempt: attemptWithoutIps},
			want:      attemptWithoutIps,
			wantErr:   false,
		},
		{
			name:      "Create an attempt without User-Agent",
			prepQuery: testUserInsert,
			args:      args{attempt: attemptWithoutUserAgent},
			want:      attemptWithoutUserAgent,
			wantErr:   false,
		},
		{
			name:      "Create an attempt with malformed IPv4",
			prepQuery: testUserInsert,
			args:      args{attempt: attemptWithMalformedIpv4},
			want:      attemptWithMalformedIpv4,
			wantErr:   true,
		},
		{
			name:      "Create an attempt with malformed IPv6",
			prepQuery: testUserInsert,
			args:      args{attempt: attemptWithMalformedIpv6},
			want:      attemptWithMalformedIpv6,
			wantErr:   true,
		},
	}
	db := getDb(t)
	defer db.Close()
	for _, testCase := range tests {
		cleanAndPopulateDb(db, testCase.prepQuery, t)
		t.Run(testCase.name, func(t *testing.T) {
			dao := LoginAttemptPgDao{}
			got, err := dao.Create(testCase.args.attempt)
			if testCase.wantErr {
				require.Error(t, err)
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.want.User.Email, got.User.Email)
				require.Equal(t, testCase.want.User.Password, got.User.Password)
				require.Equal(t, testCase.want.UserAgent, got.UserAgent)
				require.Equal(t, testCase.want.IPv4, got.IPv4)
				require.Equal(t, testCase.want.IPv6, got.IPv6)
				require.Equal(t, testCase.want.Successful, got.Successful)
			}
		})
	}
}
