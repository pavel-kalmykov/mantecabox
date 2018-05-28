package database

import (
	"mantecabox/utilities"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"log"
	"os"
	"testing"
)

func TestMigrate(t *testing.T) {
	os.Setenv("MANTECABOX_CONFIG_FILE", "configuration.test.json")
	utilities.StartDockerPostgresDb()
	db, err := GetDbReadingConfig()
	require.NoError(t, err)
	defer db.Close()

	instance, err := postgres.WithInstance(db, &postgres.Config{})
	require.NoError(t, err)

	// "file://" con dos barras es ruta relativa; con 3, absoluta
	fsrc, err := (&file.File{}).Open("file://migrations")
	require.NoError(t, err)

	m, err := migrate.NewWithInstance("file", fsrc, "postgres", instance)
	require.NoError(t, err)

	// Migrate all the way up ...
	err = m.Up()
	if err == migrate.ErrNoChange {
		log.Println("No migrations were made")
	} else if err != nil {
		log.Println("Some error ocurred: ", err)
	}
	version, dirty, err := m.Version()
	require.NoError(t, err)
	isCleanVersion := "clean"
	if dirty {
		isCleanVersion = "dirty"
	}
	log.Printf("Current migration version: %v [%v]\n", version, isCleanVersion)
}
