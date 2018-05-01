package main

import (
	"github.com/paveltrufi/mantecabox/utilities"
	"github.com/paveltrufi/mantecabox/webservice"
)

func main() {
	conf := utilities.GetServerConfiguration()
	r := webservice.Server()
	r.RunTLS(":"+conf.Port, conf.Cert, conf.Key)
}
