package database

import (
	"database/sql"
	"fmt"

	"mantecabox/models"
	"mantecabox/utilities"

	_ "github.com/lib/pq"
)

func GetPgDb() (*sql.DB, error) {
	return withConfigReaden(func(config models.Configuration) (*sql.DB, error) {
		return getDb("postgres", config)
	})
}

func GetDbReadingConfig() (*sql.DB, error) {
	return withConfigReaden(func(config models.Configuration) (*sql.DB, error) {
		return getDb(config.Engine, config)
	})
}

func withConfigReaden(f func(configuration models.Configuration) (*sql.DB, error)) (*sql.DB, error) {
	config, err := utilities.GetConfiguration()
	if err != nil {
		return nil, err
	}
	return f(config)
}

func getDb(engine string, config models.Configuration) (*sql.DB, error) {
	connectionString := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=require",
		engine, config.User, config.Password, config.Server, config.Port, config.Database)
	db, err := sql.Open(engine, connectionString)
	return db, err
}
