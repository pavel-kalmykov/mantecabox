package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/paveltrufi/mantecabox/utilities"
	"log"
)

func get() *sql.DB {
	config, err := utilities.GetConfiguration()
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require", config.User, config.Password, config.Server, config.Port, config.Database)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	return db
}
