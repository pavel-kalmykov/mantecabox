package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"context"
	"time"
	"io"
	"encoding/json"
	"path/filepath"

	"github.com/paveltrufi/mantecabox/models"
	"github.com/paveltrufi/mantecabox/utilities"
	"strings"
	"bytes"
)

type Resp struct {
	Status  string `json:"status-code"`
	Message string `json:"message"`
}

func response(w io.Writer, statusCode string, msg string) {
	r := Resp{Status: statusCode, Message: msg}
	fmt.Print(r)

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(&r)

	w.Write(b.Bytes())
}

func emptyFields(field int) string {
	responseText := "Empty body, {field} field required"
	switch field {
	case 1:
		return strings.Replace(responseText, "field", "username", 1)
	case 2:
		return strings.Replace(responseText, "field", "password", 1)
	case 3:
		return strings.Replace(responseText, "field", "username, password", 1)
	}
	return responseText
}

func compruebaUserTest(usuario models.User) (string, string) {
	testUser := "testuser"
	passUser := "testsecret"

	testOK := "Test user OK"
	testInvalid := "Invalid test user"

	if usuario.Username == testUser && usuario.Password == passUser {
		return "200", testOK
	}

	if usuario.Username == "" && usuario.Password == "" {
		return "400", emptyFields(3)
	}

	if usuario.Username == "" {
		return "400", emptyFields(1)
	}

	if usuario.Password == "" {
		return "400", emptyFields(2)
	}

	return "200", testInvalid
}

func login(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {

		usuario := models.User{}

		json.NewDecoder(req.Body).Decode(&usuario)

		w.Header().Set("Content-Type", "application/json")

		statusCode, msg := compruebaUserTest(usuario)

		response(w, statusCode, msg)
	} else if req.Method == "GET" || req.Method == "HEAD" {

	} else {
		response(w, "405", "Method Not Allowed")
	}
}

func server() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	config := utilities.GetServerConfiguration()

	mux := http.NewServeMux()
	mux.Handle("/login", http.HandlerFunc(login))

	srv := &http.Server{Addr: ":" + config.Port, Handler: mux}

	fmt.Println("Server run in port: " + config.Port + ", and IP: " + utilities.GetIPAddress())

	go func() {
		if err := srv.ListenAndServeTLS(config.Certificates.Cert, config.Certificates.Key); err != nil {
			ex, err := os.Executable()
			if err != nil {
				panic(err)
			}
			exPath := filepath.Dir(ex)
			fmt.Println(exPath)
			fmt.Println("Error ssl")
		}
	}()

	<-stopChan
	log.Println("Shutting down server...")

	ctx, fnc := context.WithTimeout(context.Background(), 5*time.Second)
	fnc()
	srv.Shutdown(ctx)

	log.Println("Server stopped correctly")
}

func main() {
	server()
}
