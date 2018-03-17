package database

import (
	"database/sql"
	"fmt"
	"github.com/paveltrufi/mantecabox/utilities"
	"log"
)

func GetDbReadingConfig() (*sql.DB, error) {
	config, err := utilities.GetConfiguration()
	if err != nil {
		log.Fatalln(err)
	}
	connectionString := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=require",
		config.Engine, config.User, config.Password, config.Server, config.Port, config.Database)
	db, err := sql.Open(config.Engine, connectionString)
	return db, err
}
