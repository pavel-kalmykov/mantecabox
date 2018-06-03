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

	testDatabaseManager := utilities.NewDatabaseManager(&config.Database)
	err = testDatabaseManager.StartDockerPostgresDb()
	if err != nil {
		logrus.Fatal("Unable to start Docker: " + err.Error())
	}
	err = testDatabaseManager.RunMigrations()
	if err != nil {
		logrus.Fatal("Unable to run migrations: " + err.Error())
	}

	r := webservice.Router(true, &config)
	if r == nil {
		logrus.Fatal(fmt.Sprintf("Unable to start web server: %v", err))
		return
	}
	r.RunTLS(fmt.Sprintf(":%v", config.Server.Port), config.Server.Cert, config.Server.Key)
}
