package main

import (
	"log"

	"github.com/zalando/go-keyring"
)

func main() {
	service := "mantecabox"
	user := "paveltrufi"
	// password := "secret"

	// set password
	// err := keyring.Set(service, user, password)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// get password
	secret, err := keyring.Get(service, user)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(secret)
}
