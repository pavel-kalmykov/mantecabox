package webservice

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/paveltrufi/mantecabox/controllers"
	"github.com/paveltrufi/mantecabox/models"
	"github.com/paveltrufi/mantecabox/utilities"
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

func compruebaUserTest(usuario models.User) (int, string) {
	const testUser = "testuser"
	const passUser = "testsecret"
	const testOK = "Test user OK"
	const testInvalid = "Invalid test user"

	if usuario.Username == testUser && usuario.Password == passUser {
		return http.StatusOK, testOK
	}

	if usuario.Username == "" && usuario.Password == "" {
		return http.StatusBadRequest, replaceEmptyField("username, password")
	}

	if usuario.Username == "" {
		return http.StatusBadRequest, replaceEmptyField("username")
	}

	if usuario.Password == "" {
		return http.StatusBadRequest, replaceEmptyField("password")
	}

	return http.StatusOK, testInvalid
}

func login(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		usuario := models.User{}
		json.NewDecoder(req.Body).Decode(&usuario)
		statusCode, msg := compruebaUserTest(usuario)
		response(w, statusCode, msg)
	case http.MethodGet, http.MethodHead:
		break
	default:
		response(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}

func server() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	config := utilities.GetServerConfiguration()
	mux := http.NewServeMux()
	mux.Handle("/login", http.HandlerFunc(login))
	srv := &http.Server{Addr: ":" + config.Port, Handler: mux}
	log.Println("Server run in port: " + config.Port + ", and IP: " + utilities.GetIPAddress())

	go func() {
		if err := srv.ListenAndServeTLS(config.Certificates.Cert, config.Certificates.Key); err != nil {
			ex, err := os.Executable()
			if err != nil {
				panic(err)
			}
			exPath := filepath.Dir(ex)
			log.Println(exPath)
		}
	}()

	<-stopChan

	log.Println("Shutting down server...")
	ctx, fnc := context.WithTimeout(context.Background(), 5*time.Second)
	fnc()
	srv.Shutdown(ctx)
	log.Println("Server stopped correctly")
}
