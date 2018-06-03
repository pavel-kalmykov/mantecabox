package webservice

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"mantecabox/database"
	"mantecabox/services"
	"mantecabox/utilities/aes"

	"github.com/appleboy/gofight"
	"github.com/buger/jsonparser"
	"github.com/gin-gonic/gin"
	"github.com/go-http-utils/headers"
	"gopkg.in/dgrijalva/jwt-go.v3"

	"mantecabox/dao/factory"
	"mantecabox/models"
	"mantecabox/utilities"

	"github.com/stretchr/testify/require"
)

// base64(sha512(password))
const (
	correctPassword   = "MTg2NEU3NTRCN0QyOENDOTk0OURDQkI1MEVFM0FFNEY3NTdCRjc1MTAwRjBDMkMzRTM3RDUxQ0Y0QURDNEVDREU0NDhCODQ2ODdEQTg3QjY5RTJGNkRCNTQwRUVFODMwNDM1MjY0RDlGNDcwNzc5MTQ4MUYyNUQ0NUUyOEQ5MTA="
	testUserEmail     = "testuser@example.com"
	testUserRealEmail = "mantecabox@gmail.com"
	modifiedUserEmail = "modifieduser@example.com"
)

var (
	userDao         = factory.UserDaoFactory("postgres")
	secureRouter    = Router(true)
	router          = Router(false)
	tokenParserFunc = func(token *jwt.Token) (interface{}, error) {
		return aes.Key, nil
	}
)

type subtest struct {
	name string
	test func(*testing.T)
}

type authResponse struct {
	Code   int    `json:"code"`
	Expire string `json:"expire"`
	Token  string `json:"token"`
}

func TestMain(m *testing.M) {
	utilities.StartDockerPostgresDb()

	code := m.Run()

	db, err := database.GetPgDb()
	if err == nil {
		cleanDb(db)
	}
	os.Exit(code)
}

func TestGetUsers(t *testing.T) {
	db := getDb(t)
	defer db.Close()
	tests := []subtest{
		{
			name: "When there is at least one user, you retrieve it in an array",
			test: func(t *testing.T) {
				userDao.Create(&models.User{Credentials: models.Credentials{Email: testUserEmail, Password: "testpassword"}})
				r := gofight.New()
				r.GET("/users").
					SetDebug(true).
					Run(router, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
					require.Equal(t, http.StatusOK, res.Code)
					expected, err := json.Marshal([]map[string]string{{"email": testUserEmail}})
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
					Run(router, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
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
				r.GET("/users/" + testUserEmail).
					SetDebug(true).
					Run(router, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
					require.Equal(t, http.StatusOK, res.Code)
					expected, err := json.Marshal(map[string]string{"email": testUserEmail})
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
					Run(router, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
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
		userDao.Create(&models.User{Credentials: models.Credentials{Email: testUserEmail, Password: "testpassword"}})
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
				r.POST("/register").
					SetDebug(true).
					SetJSON(gofight.D{
					"email":    testUserEmail,
					"password": correctPassword,
				}).
					Run(router, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
					require.Equal(t, http.StatusCreated, res.Code)
					expected, err := json.Marshal(map[string]string{"email": testUserEmail})
					require.NoError(t, err)
					require.JSONEq(t, string(expected), res.Body.String())
				})
			},
		},
		{
			name: "When you send malformed data to the service (non-hashed password), it returns the proper error",
			test: func(t *testing.T) {
				r := gofight.New()
				r.POST("/register").
					SetDebug(true).
					SetJSON(gofight.D{
					"email":    testUserEmail,
					"password": "bWFudGVjYWJveA==",
				}).
					Run(router, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
					require.Equal(t, http.StatusBadRequest, res.Code)
					expected, err := json.Marshal(map[string]string{
						"message": "Unable to register user: " + services.InvalidPasswordError.Error(),
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
				r.POST("/register").
					SetDebug(true).
					SetBody("{{,invent,}}").
					Run(router, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
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
				r.PUT("/users/" + testUserEmail).
					SetJSON(gofight.D{
					"email":    modifiedUserEmail,
					"password": correctPassword,
				}).
					Run(router, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
					require.Equal(t, http.StatusCreated, res.Code)
					expected, err := json.Marshal(map[string]string{
						"email": modifiedUserEmail,
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
					"email":    modifiedUserEmail,
					"password": correctPassword,
				}).
					Run(router, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
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
					"email":    testUserEmail,
					"password": "bWFudGVjYWJveA==",
				}).
					Run(router, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
					require.Equal(t, http.StatusBadRequest, res.Code)
					expected, err := json.Marshal(map[string]string{
						"message": "Unable to modify user: " + services.InvalidPasswordError.Error(),
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
					Run(router, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
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
		userDao.Create(&models.User{Credentials: models.Credentials{Email: testUserEmail, Password: "testpassword"}})
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
				r.DELETE("/users/" + testUserEmail).
					SetDebug(true).
					Run(router, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
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
					Run(router, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
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
		userDao.Create(&models.User{Credentials: models.Credentials{Email: testUserEmail, Password: "testpassword"}})
		t.Run(tt.name, tt.test)
	}
}

func TestJWTRouter(t *testing.T) {
	db := getDb(t)
	defer db.Close()
	tests := []subtest{
		{
			name: "When you try to login with correct credentials, receive a new token",
			test: func(t *testing.T) {
				user, err := userDao.GetByPk(testUserEmail)
				require.NoError(t, err)
				r := gofight.New()
				r.POST("/login").
					SetDebug(true).
					SetQuery(gofight.H{
					"verification_code": user.TwoFactorAuth.ValueOrZero(),
				}).
					SetJSON(gofight.D{
					"username": testUserEmail,
					"password": correctPassword,
				}).
					Run(secureRouter, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
					require.Equal(t, http.StatusOK, res.Code)

					bodyBytes := res.Body.Bytes()
					code, err := jsonparser.GetInt(bodyBytes, "code")
					require.NoError(t, err)
					expireString, err := jsonparser.GetString(bodyBytes, "expire")
					require.NoError(t, err)
					expireDate, err := time.Parse(time.RFC3339, expireString)
					require.NoError(t, err)
					tokenString, err := jsonparser.GetString(bodyBytes, "token")
					require.NoError(t, err)
					token, _ := jwt.Parse(tokenString, tokenParserFunc)

					require.EqualValues(t, http.StatusOK, code)
					require.True(t, expireDate.After(time.Now().Local().Add(time.Hour-time.Minute)))
					require.True(t, expireDate.Before(time.Now().Local().Add(time.Hour)))
					require.True(t, token.Valid)
				})
			},
		},
		{
			name: "When you try to access a protected route with a token, allow it",
			test: func(t *testing.T) {
				performActionWithToken(t, func(auth authResponse) {
					r := gofight.New()
					r.GET("/users").
						SetDebug(true).
						SetHeader(gofight.H{
						headers.Authorization: "Bearer " + auth.Token,
					}).Run(secureRouter, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						require.Equal(t, http.StatusOK, res.Code)
					})
				})
			},
		},
		{
			name: "When you try to access a protected route without a token, deny it",
			test: func(t *testing.T) {
				r := gofight.New()
				r.GET("/users").
					SetDebug(true).
					Run(secureRouter, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
					require.Equal(t, http.StatusUnauthorized, res.Code)
					expected, err := json.Marshal(map[string]interface{}{
						"code":    401,
						"message": "auth header is empty",
					})
					require.NoError(t, err)
					require.JSONEq(t, string(expected), res.Body.String())
				})
			},
		},
		{
			name: "When you try to login without the 2FA verification code, deny it",
			test: func(t *testing.T) {
				r := gofight.New()
				r.POST("/login").
					SetDebug(true).
					SetJSON(gofight.D{
					"username": testUserEmail,
					"password": correctPassword,
				}).
					Run(secureRouter, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
					require.Equal(t, http.StatusUnauthorized, res.Code)
					expected, err := json.Marshal(map[string]interface{}{
						"code":    401,
						"message": "incorrect Username, Password or Verification Code",
					})
					require.NoError(t, err)
					require.JSONEq(t, string(expected), res.Body.String())
				})
			},
		},
		{
			name: "When you try to access a protected route with a corrupt token, deny it",
			test: func(t *testing.T) {
				performActionWithToken(t, func(auth authResponse) {
					auth.Token += "jfklasdjflakjñs"
					r := gofight.New()
					r.GET("/users").
						SetDebug(true).
						SetHeader(gofight.H{
						headers.Authorization: "Bearer " + auth.Token,
					}).Run(secureRouter, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						require.Equal(t, http.StatusUnauthorized, res.Code)

						expectedCode, err := jsonparser.GetInt(res.Body.Bytes(), "code")
						require.NoError(t, err)
						expectedMessage, err := jsonparser.GetString(res.Body.Bytes(), "message")
						require.NoError(t, err)

						require.EqualValues(t, expectedCode, http.StatusUnauthorized)
						require.True(t, strings.HasPrefix(string(expectedMessage), "illegal base64 data at input byte "))
					})
				})
			},
		},
		{
			name: "When you refresh a token, its expiring date is after the original one, but both still valid",
			test: func(t *testing.T) {
				performActionWithToken(t, func(auth authResponse) {
					time.Sleep(time.Second)
					r := gofight.New()
					r.GET("/refresh-token").
						SetDebug(true).
						SetHeader(gofight.H{
						headers.Authorization: "Bearer " + auth.Token,
					}).Run(secureRouter, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
						var refreshedAuth authResponse
						err := json.Unmarshal(res.Body.Bytes(), &refreshedAuth)
						require.NoError(t, err)

						expireDate, err := time.Parse(time.RFC3339, auth.Expire)
						require.NoError(t, err)
						refreshedExpireDate, err := time.Parse(time.RFC3339, refreshedAuth.Expire)
						require.True(t, expireDate.Before(refreshedExpireDate))
					})
				})
			},
		},
	}
	for _, tt := range tests {
		cleanDb(db)
		user, err := services.RegisterUser(&models.Credentials{
			Email:    testUserEmail,
			Password: correctPassword,
		})
		require.NoError(t, err)
		_, err = services.Generate2FACodeAndSaveToUser(&user)
		require.NoError(t, err)
		t.Run(tt.name, tt.test)
	}
}

func TestGenerate2FAAndSendMail(t *testing.T) {
	db := getDb(t)
	defer db.Close()
	tests := []subtest{
		{
			name: "When you generate a 2FA code to an existent user with a valid email, the code is generated and the mail sent",
			test: func(t *testing.T) {
				r := gofight.New()
				r.POST("/2fa-verification").
					SetDebug(true).
					SetJSON(gofight.D{
					"email":    testUserRealEmail,
					"password": correctPassword,
				}).
					Run(secureRouter, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
					require.Equal(t, http.StatusOK, res.Code)
					expected, err := json.Marshal(map[string]interface{}{
						"message": "Verification code sent correctly to " + testUserRealEmail + ". Check your inbox!",
					})
					require.NoError(t, err)
					require.JSONEq(t, string(expected), res.Body.String())
				})
			},
		},
		{
			name: "When you generate a 2FA code to an unexistent user with a valid email, send a not found error",
			test: func(t *testing.T) {
				r := gofight.New()
				r.POST("/2fa-verification").
					SetDebug(true).
					SetJSON(gofight.D{
					"email":    testUserEmail,
					"password": correctPassword,
				}).
					Run(secureRouter, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
					require.Equal(t, http.StatusNotFound, res.Code)
					expected, err := json.Marshal(map[string]interface{}{
						"message": "Wrong credentials for: " + testUserEmail + ". Please check the username and password are correct!",
					})
					require.NoError(t, err)
					require.JSONEq(t, string(expected), res.Body.String())
				})
			},
		},
	}
	for _, tt := range tests {
		cleanDb(db)
		user, err := services.RegisterUser(&models.Credentials{
			Email:    testUserRealEmail,
			Password: correctPassword,
		})
		require.NoError(t, err)
		_, err = services.Generate2FACodeAndSaveToUser(&user)
		require.NoError(t, err)
		t.Run(tt.name, tt.test)
	}

}

func performActionWithToken(t *testing.T, action func(response authResponse)) {
	performActionWithTokenAndCustomRouter(t, secureRouter, action)
}

func performActionWithTokenAndCustomRouter(t *testing.T, customRouter *gin.Engine, action func(response authResponse)) {
	user, err := userDao.GetByPk(testUserEmail)
	require.NoError(t, err)
	r := gofight.New()
	r.POST("/login").
		SetDebug(true).
		SetQuery(gofight.H{
		"verification_code": user.TwoFactorAuth.ValueOrZero(),
	}).
		SetJSON(gofight.D{
		"username": testUserEmail,
		"password": correctPassword,
	}).
		Run(customRouter, func(res gofight.HTTPResponse, req gofight.HTTPRequest) {
		require.Equal(t, http.StatusOK, res.Code)

		var auth authResponse
		err := json.Unmarshal(res.Body.Bytes(), &auth)
		require.NoError(t, err)
		action(auth)
	})
}

func getDb(t *testing.T) *sql.DB {
	db, err := database.GetPgDb()
	require.NoError(t, err)
	require.NotNil(t, db)
	return db
}

func cleanDb(db *sql.DB) {
	db.Exec("DELETE FROM users")
}
