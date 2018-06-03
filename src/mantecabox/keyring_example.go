package main

import (
	"log"

	"github.com/zalando/go-keyring"
)

func main() {

	service := "mantecabox"
	user := "token_response"

	secret, err := keyring.Get(service, user)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(secret)
}
