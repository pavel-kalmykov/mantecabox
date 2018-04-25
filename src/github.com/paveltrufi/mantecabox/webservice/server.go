package webservice

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/paveltrufi/mantecabox/utilities"

	"github.com/paveltrufi/mantecabox/models"
)

type Resp struct {
	Message string `json:"message"`
}

func response(w http.ResponseWriter, statusCode int, msg string) {
	response := Resp{Message: msg}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(&response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(b.Bytes())
}

func replaceEmptyField(field string) string {
	const responseText = "Empty body, {field} field required"
	return strings.Replace(responseText, "field", field, 1)
}

func compruebaUserTest(credentials models.Credentials) (int, string) {
	const testUser = "testuser"
	const passUser = "testsecret"
	const testOK = "Test user OK"
	const testInvalid = "Invalid test user"

	if credentials.Username == testUser && credentials.Password == passUser {
		return http.StatusOK, testOK
	}

	if credentials.Username == "" && credentials.Password == "" {
		return http.StatusBadRequest, replaceEmptyField("username, password")
	}

	if credentials.Username == "" {
		return http.StatusBadRequest, replaceEmptyField("username")
	}

	if credentials.Password == "" {
		return http.StatusBadRequest, replaceEmptyField("password")
	}

	return http.StatusOK, testInvalid
}

func login(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodPost:
		var usuario models.Credentials
		err := c.ShouldBindJSON(&usuario)
		if err != nil {
			panic(err)
		}
		statusCode, msg := compruebaUserTest(usuario)
		response(c.Writer, statusCode, msg)
	case http.MethodGet, http.MethodHead:
		break
	default:
		response(c.Writer, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}

func Server() {
	conf := utilities.GetServerConfiguration()
	r := gin.Default()
	r.Any("/login", login)
	r.RunTLS(":"+conf.Port, conf.Cert, conf.Key)
}
