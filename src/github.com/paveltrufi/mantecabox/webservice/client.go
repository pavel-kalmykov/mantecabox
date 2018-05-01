package webservice

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"os"

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
	usuario := models.User{Credentials: models.Credentials{Username: testUser, Password: passUser}}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(usuario)

	r, err := client.Post(endpoint, "application/json", b) // enviamos por POST

	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, r.Body)
}
