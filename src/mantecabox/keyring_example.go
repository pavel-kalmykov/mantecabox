package main

import (
	"log"

	"github.com/zalando/go-keyring"
)

func main() {

	service := "mantecabox"
	user := "raul_pairo@icloud.com"

	secret, err := keyring.Get(service, user)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(secret)
}
