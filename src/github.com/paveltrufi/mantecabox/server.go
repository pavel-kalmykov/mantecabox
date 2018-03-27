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
)

func chk(e error) {
	if e != nil {
		panic(e)
	}
}

type resp struct {
	Ok  bool
	message string
}

func response(w io.Writer, ok bool, msg string) {
	r := resp{Ok: ok, message: msg}
	rJSON, err := json.Marshal(&r)
	chk(err)
	w.Write(rJSON)
}

func emptyFields() string {
	responseText := "'Empty body', '{username, password}' field required'"
	return responseText
}

func compruebaUserTest(user string, password string) (bool, string) {
	testUser := "testuser"
	passUser := "testsecret"

	if user == testUser {
		if password == passUser {
			return true, "Test user OK"
		} else if password == "" {
			return false, emptyFields()
		} else {
			return true, "Invalid test user"
		}
	} else if password == "" {
		return false, emptyFields()
	} else {
		return true, "Invalid test user"
	}
}

func login(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		req.ParseForm()
		w.Header().Set("Content-Type", "text/plain")

		username := req.Form.Get("username")
		password := req.Form.Get("password")

		ok, msg := compruebaUserTest(username, password)
		response(w, ok, msg)
	} else if req.Method == "GET" || req.Method == "HEAD" {

	} else {
		w.WriteHeader(405)
		w.Write([]byte(`{"message": "Method Not Allowed"}`))
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
