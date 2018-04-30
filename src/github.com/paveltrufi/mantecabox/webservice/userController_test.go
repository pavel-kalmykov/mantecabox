package webservice

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/appleboy/gofight"
	"github.com/buger/jsonparser"
	"github.com/paveltrufi/mantecabox/services"

	"github.com/paveltrufi/mantecabox/dao/factory"
	"github.com/paveltrufi/mantecabox/dao/postgres"
	"github.com/paveltrufi/mantecabox/models"
	"github.com/paveltrufi/mantecabox/utilities"
	"github.com/stretchr/testify/require"
)

var userDao = factory.UserDaoFactory("postgres")

type subtest struct {
	name string
	test func(*testing.T)
}

func TestMain(m *testing.M) {
	utilities.StartDockerPostgresDb()
	// os.Setenv("MANTECABOX_CONFIG_FILE", "configuration.test.json")
	// it must be initialized from the run script configuration

	code := m.Run()

	db := postgres.GetPgDb()
	cleanDb(db)
	os.Exit(code)
}

func TestGetUsers(t *testing.T) {
	db := getDb(t)
	defer db.Close()
	tests := []subtest{
		{
			name: "When there is at least one user, you retrieve it in an array",
			test: func(t *testing.T) {
				userDao.Create(&models.User{Credentials: models.Credentials{Username: "testuser", Password: "testpassword"}})
				r := gofight.New()
				r.GET("/users").
					SetDebug(true).
					Run(Server(), func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						require.Equal(t, http.StatusOK, res.Code)
						expected, err := json.Marshal([]map[string]string{{"username": "testuser"}})
						require.NoError(t, err)
						require.JSONEq(t, string(expected), res.Body.String())
					})
			},
		},
		{
			name: "When there are no users in the database, return an empty array",
			test: func(t *testing.T) {
				r := gofight.New()
				r.GET("/users").
					SetDebug(true).
					Run(Server(), func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						require.Equal(t, http.StatusOK, res.Code)
						expected, err := json.Marshal([]map[string]string{})
						require.NoError(t, err)
						require.JSONEq(t, string(expected), res.Body.String())
					})
			},
		},
	}
	for _, tt := range tests {
		cleanDb(db)
		t.Run(tt.name, tt.test)
	}
}

func TestGetUser(t *testing.T) {
	db := getDb(t)
	defer db.Close()
	tests := []subtest{
		{
			name: "When you pass an existent user, retrieve it",
			test: func(t *testing.T) {
				r := gofight.New()
				r.GET("/users/testuser").
					SetDebug(true).
					Run(Server(), func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						require.Equal(t, http.StatusOK, res.Code)
						expected, err := json.Marshal(map[string]string{"username": "testuser"})
						require.NoError(t, err)
						require.JSONEq(t, string(expected), res.Body.String())
					})
			},
		},
		{
			name: "When you pass a non-existent user, send error",
			test: func(t *testing.T) {
				r := gofight.New()
				r.GET("/users/nonexistent").
					SetDebug(true).
					Run(Server(), func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						require.Equal(t, http.StatusNotFound, res.Code)
						expected, err := json.Marshal(map[string]string{"message": "Unable to find user: nonexistent"})
						require.NoError(t, err)
						require.JSONEq(t, string(expected), res.Body.String())
					})
			},
		},
	}
	for _, tt := range tests {
		cleanDb(db)
		userDao.Create(&models.User{Credentials: models.Credentials{Username: "testuser", Password: "testpassword"}})
		t.Run(tt.name, tt.test)
	}
}

func TestRegisterUser(t *testing.T) {
	db := getDb(t)
	defer db.Close()
	tests := []subtest{
		{
			name: "When you send some new proper credentials (valid username, password in SHA512), register a new user properly",
			test: func(t *testing.T) {
				r := gofight.New()
				r.POST("/users").
					SetDebug(true).
					SetJSON(gofight.D{
						"username": "testuser",
						"password": "MTg2NEU3NTRCN0QyOENDOTk0OURDQkI1MEVFM0FFNEY3NTdCRjc1MTAwRjBDMkMzRTM3RDUxQ0Y0QURDNEVDREU0NDhCODQ2ODdEQTg3QjY5RTJGNkRCNTQwRUVFODMwNDM1MjY0RDlGNDcwNzc5MTQ4MUYyNUQ0NUUyOEQ5MTA=",
					}).
					Run(Server(), func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						require.Equal(t, http.StatusCreated, res.Code)
						expected, err := json.Marshal(map[string]string{"username": "testuser"})
						require.NoError(t, err)
						require.JSONEq(t, string(expected), res.Body.String())
					})
			},
		},
		{
			name: "When you send malformed data to the service (non-hashed password), it returns the proper error",
			test: func(t *testing.T) {
				r := gofight.New()
				r.POST("/users").
					SetDebug(true).
					SetJSON(gofight.D{
						"username": "testuser",
						"password": "bWFudGVjYWJveA==",
					}).
					Run(Server(), func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						require.Equal(t, http.StatusBadRequest, res.Code)
						expected, err := json.Marshal(map[string]string{
							"message": "Unable to register user: " + services.InvalidPasswordError,
						})
						require.NoError(t, err)
						require.JSONEq(t, string(expected), res.Body.String())
					})
			},
		},
		{
			name: "When you send a malformed JSON, it returns a parsing error",
			test: func(t *testing.T) {
				r := gofight.New()
				r.POST("/users").
					SetDebug(true).
					SetBody("{{,invent,}}").
					Run(Server(), func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						require.Equal(t, http.StatusBadRequest, res.Code)
						expected, err := jsonparser.GetString(res.Body.Bytes(), "message")
						require.NoError(t, err)
						require.Contains(t, expected, "Unable to parse JSON:")
					})
			},
		},
	}
	for _, tt := range tests {
		cleanDb(db)
		t.Run(tt.name, tt.test)
	}
}

func TestModifyUser(t *testing.T) {
	db := getDb(t)
	defer db.Close()
	tests := []subtest{
		{
			name: "When you modify an existent user with proper data, modify it",
			test: func(t *testing.T) {
				r := gofight.New()
				r.PUT("/users/testuser").
					SetJSON(gofight.D{
						"username": "modifiedUser",
						"password": "MTg2NEU3NTRCN0QyOENDOTk0OURDQkI1MEVFM0FFNEY3NTdCRjc1MTAwRjBDMkMzRTM3RDUxQ0Y0QURDNEVDREU0NDhCODQ2ODdEQTg3QjY5RTJGNkRCNTQwRUVFODMwNDM1MjY0RDlGNDcwNzc5MTQ4MUYyNUQ0NUUyOEQ5MTA=",
					}).
					Run(Server(), func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						require.Equal(t, http.StatusCreated, res.Code)
						expected, err := json.Marshal(map[string]string{
							"username": "modifiedUser",
						})
						require.NoError(t, err)
						require.JSONEq(t, string(expected), res.Body.String())
					})
			},
		},
		{
			name: "When you modify a non existent user, send a not found error",
			test: func(t *testing.T) {
				r := gofight.New()
				r.PUT("/users/nonexistent").
					SetJSON(gofight.D{
						"username": "modifiedUser",
						"password": "MTg2NEU3NTRCN0QyOENDOTk0OURDQkI1MEVFM0FFNEY3NTdCRjc1MTAwRjBDMkMzRTM3RDUxQ0Y0QURDNEVDREU0NDhCODQ2ODdEQTg3QjY5RTJGNkRCNTQwRUVFODMwNDM1MjY0RDlGNDcwNzc5MTQ4MUYyNUQ0NUUyOEQ5MTA=",
					}).
					Run(Server(), func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						require.Equal(t, http.StatusNotFound, res.Code)
					})
			},
		},
		{
			name: "When you modify an existent user with malformed data (without username), return the proper error",
			test: func(t *testing.T) {
				r := gofight.New()
				r.PUT("/users/testuser").
					SetDebug(true).
					SetJSON(gofight.D{
						"username": "testuser",
						"password": "bWFudGVjYWJveA==",
					}).
					Run(Server(), func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						require.Equal(t, http.StatusBadRequest, res.Code)
						expected, err := json.Marshal(map[string]string{
							"message": "Unable to modify user: " + services.InvalidPasswordError,
						})
						require.NoError(t, err)
						require.JSONEq(t, string(expected), res.Body.String())
					})
			},
		},
		{
			name: "When you send a malformed JSON, it returns a parsing error",
			test: func(t *testing.T) {
				r := gofight.New()
				r.PUT("/users/testuser").
					SetDebug(true).
					SetBody("{{,invent,}}").
					Run(Server(), func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						require.Equal(t, http.StatusBadRequest, res.Code)
						expected, err := jsonparser.GetString(res.Body.Bytes(), "message")
						require.NoError(t, err)
						require.Contains(t, expected, "Unable to parse JSON:")
					})
			},
		},
	}
	for _, tt := range tests {
		cleanDb(db)
		userDao.Create(&models.User{Credentials: models.Credentials{Username: "testuser", Password: "testpassword"}})
		t.Run(tt.name, tt.test)
	}
}

func TestDeleteUser(t *testing.T) {
	db := getDb(t)
	defer db.Close()
	tests := []subtest{
		{
			name: "When you delete an existent, it gets deleted",
			test: func(t *testing.T) {
				r := gofight.New()
				r.DELETE("/users/testuser").
					SetDebug(true).
					Run(Server(), func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						require.Equal(t, http.StatusNoContent, res.Code)
						require.Empty(t, res.Body)
					})
			},
		},
		{
			name: "When you delete a non existent, return a not found message",
			test: func(t *testing.T) {
				r := gofight.New()
				r.DELETE("/users/nonexistent").
					SetDebug(true).
					Run(Server(), func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						require.Equal(t, http.StatusNotFound, res.Code)
						expected, err := json.Marshal(map[string]string{
							"message": "Unable to find user: nonexistent",
						})
						require.NoError(t, err)
						require.JSONEq(t, string(expected), res.Body.String())
					})
			},
		},
	}
	for _, tt := range tests {
		cleanDb(db)
		userDao.Create(&models.User{Credentials: models.Credentials{Username: "testuser", Password: "testpassword"}})
		t.Run(tt.name, tt.test)
	}
}

func getDb(t *testing.T) *sql.DB {
	db := postgres.GetPgDb()
	require.NotNil(t, db)
	return db
}

func cleanDb(db *sql.DB) {
	db.Exec("DELETE FROM users")
}
