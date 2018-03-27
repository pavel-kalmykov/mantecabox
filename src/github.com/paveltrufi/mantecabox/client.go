package main

import (
	"net/http"
	"crypto/tls"
	"net/url"
	"io"
	"os"
	"fmt"
)

func client() {

	endpoint := "https://localhost:10443/login/"

	/* creamos un cliente especial que no comprueba la validez de los certificados
	esto es necesario por que usamos certificados autofirmados (para pruebas) */
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// ** ejemplo de registro
	data := url.Values{}             // estructura para contener los valores
	data.Set("username", "testuser")          // comando (string)
	data.Set("password", "testsecret") // usuario (string)

	r, err := client.PostForm(endpoint, data) // enviamos por POST

	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, r.Body) // mostramos el cuerpo de la respuesta (es un reader)
	fmt.Println()
}

func main() {
	client()
}
