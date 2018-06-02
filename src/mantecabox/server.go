package main

import (
	"fmt"

	"mantecabox/utilities"
	"mantecabox/webservice"

	"github.com/sirupsen/logrus"
)

func main() {
	config, err := utilities.GetConfiguration()
	if err != nil {
		logrus.Fatal(fmt.Sprintf("Unable to read configuration file: %v", err))
		return
	}
	r := webservice.Router(true)
	r.RunTLS(fmt.Sprintf(":%v", config.Server.Port), config.Server.Cert, config.Server.Key)
}
