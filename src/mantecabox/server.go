package main

import (
	"mantecabox/utilities"
	"mantecabox/webservice"
)

func main() {
	conf := utilities.GetServerConfiguration()
	r := webservice.Router(true)
	r.RunTLS(":"+conf.Port, conf.Cert, conf.Key)
}
