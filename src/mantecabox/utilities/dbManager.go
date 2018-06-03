package utilities

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"mantecabox/models"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const dockerContainerName = "sds-postgres"

type (
	DatabaseManager interface {
		StartDockerPostgresDb() error
		RunMigrations() error
	}

	DatabaseManagerImpl struct {
		database *models.Database
	}
)

func NewDatabaseManager(database *models.Database) DatabaseManager {
	if database == nil {
		return nil
	}
	return DatabaseManagerImpl{
		database: database,
	}
}

// startDockerPostgresDb ejecuta un comando para comprobar si el contenedor de la base de datos está en ejecución o no.
// Este comando devolverá "true\n" o "false\n", así que comprobamos que si no devuelve true lo iniciemos (esto lo
// sabremos si al ejecutarse el comando éste ha devuelto el nombre del mismo seguido de un salto de línea).
func (databaseManager DatabaseManagerImpl) StartDockerPostgresDb() error {
	command := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", dockerContainerName)
	output, err := command.Output()
	if err != nil {
		return err
	}
	// usamos hasPrefix para no tener que controlar los saltos de línea
	if !(strings.HasPrefix(string(output), "true")) {
		output, err := exec.Command("docker", "container", "start", dockerContainerName).Output()
		if err != nil {
			return err
		}
		if strings.HasPrefix(string(output), dockerContainerName) {
			logrus.Debug("Docker container '%s' started\n", dockerContainerName)
			time.Sleep(time.Second * 2) // Lo que tarde en arrancar, más o menos
		} else {
			return errors.New("unable to start Postgre's docker container")
		}
	} else {
		logrus.Debug("Docker container '%s' already running\n", dockerContainerName)
	}
	return nil
}

func (databaseManager DatabaseManagerImpl) RunMigrations() error {
	db, err := GetDbFromConfig(databaseManager.database)
	if err != nil {
		return err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	instance, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file:///%v/migrations", os.Getenv("GOPATH")),
		databaseManager.database.Name, driver)
	if err != nil {
		return err
	}
	err = instance.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	version, _, err := instance.Version()
	if err != nil {
		return err
	}
	logrus.Infof("Database schema up-to-date (version %v)", version)
	return nil
}
