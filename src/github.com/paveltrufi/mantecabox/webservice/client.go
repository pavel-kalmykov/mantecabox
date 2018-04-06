package main

import (
	"net/http"
	"crypto/tls"
	"io"
	"os"
	"fmt"
	"bytes"
	"encoding/json"
	"github.com/paveltrufi/mantecabox/models"
)

func client() {

	endpoint := "https://localhost:10443/login"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	usuario := models.User{Username: "testuser", Password:"testsecret"}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(usuario)

	r, err := client.Post(endpoint, "application/json; charset=utf-8", b) // enviamos por POST

	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, r.Body) // mostramos el cuerpo de la respuesta (es un reader)
	fmt.Println()
}

func main() {
	client()
}
