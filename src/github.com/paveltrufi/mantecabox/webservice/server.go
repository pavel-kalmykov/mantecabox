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
	"strings"
	"bytes"
)

type Resp struct {
	Status  bool `json:"status"`
	Message string `json:"message"`
}

func response(w io.Writer, ok bool, msg string) {
	r := Resp{Status: ok, Message: msg}
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

func compruebaUserTest(usuario models.User) (bool, string) {
	testUser := "testuser"
	passUser := "testsecret"

	testOK := "Test user OK"
	testInvalid := "Invalid test user"

	if usuario.Username == testUser && usuario.Password == passUser {
		return true, testOK
	}

	if usuario.Username == "" && usuario.Password == "" {
		return false, emptyFields(3)
	}

	if usuario.Username == "" {
		return false, emptyFields(1)
	}

	if usuario.Password == "" {
		return false, emptyFields(2)
	}

	return true, testInvalid
}

func login(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {

		usuario := models.User{}

		json.NewDecoder(req.Body).Decode(&usuario)

		w.Header().Set("Content-Type", "application/json")

		ok, msg := compruebaUserTest(usuario)

		response(w, ok, msg)
	} else if req.Method == "GET" || req.Method == "HEAD" {

	} else {
		w.WriteHeader(405)
		response(w, false, "Method Not Allowed")
	}
}

func server() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	mux := http.NewServeMux()
	mux.Handle("/login", http.HandlerFunc(login))

	srv := &http.Server{Addr: ":10443", Handler: mux}

	fmt.Println("Server on")

	go func() {
		if err := srv.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
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
