package utilities

import (
	"database/sql"
	"fmt"

	"mantecabox/models"

	_ "github.com/lib/pq"
)

func GetPgDb() (*sql.DB, error) {
	return withConfigReaden(func(config *models.Database) (*sql.DB, error) {
		return getDb("postgres", config)
	})
}

func GetDbFromConfig(configuration *models.Database) (*sql.DB, error) {
	return getDb(configuration.Engine, configuration)
}

func withConfigReaden(f func(configuration *models.Database) (*sql.DB, error)) (*sql.DB, error) {
	config, err := GetConfiguration()
	if err != nil {
		return nil, err
	}
	return f(&config.Database)
}

func getDb(engine string, config *models.Database) (*sql.DB, error) {
	connectionString := fmt.Sprintf("%v://%v:%v@%v:%v/%v?sslmode=require",
		engine, config.User, config.Password, config.Host, config.Port, config.Name)
	db, err := sql.Open(engine, connectionString)
	return db, err
}
