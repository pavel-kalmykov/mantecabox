package main

import (
	"net/http"
	"crypto/tls"
	"io"
	"os"
	"bytes"
	"encoding/json"
	"github.com/paveltrufi/mantecabox/models"
)

func client() {
	const endpoint = "https://localhost:10443/login"
	const testUser = "testuser"
	const passUser = "testsecret"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	usuario := models.User{Username: testUser, Password:passUser}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(usuario)

	r, err := client.Post(endpoint, "application/json", b) // enviamos por POST

	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, r.Body)
}

func main() {
	client()
}
