package utilities

import (
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const dockerContainerName = "sds-postgres"

// startDockerPostgresDb ejecuta un comando para comprobar si el contenedor de la base de datos está en ejecución o no.
// Este comando devolverá "true\n" o "false\n", así que comprobamos que si no devuelve true lo iniciemos (esto lo
// sabremos si al ejecutarse el comando éste ha devuelto el nombre del mismo seguido de un salto de línea).
func StartDockerPostgresDb() {
	command := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", dockerContainerName)
	output, err := command.Output()
	checkErr(err)
	// usamos hasPrefix para no tener que controlar los saltos de línea
	if !(strings.HasPrefix(string(output), "true")) {
		output, err := exec.Command("docker", "container", "start", dockerContainerName).Output()
		checkErr(err)
		if strings.HasPrefix(string(output), dockerContainerName) {
			logrus.Debug("Docker container '%s' started\n", dockerContainerName)
			time.Sleep(time.Second * 2) // Lo que tarde en arrancar, más o menos
		} else {
			panic("Unable to start Postgre's docker container!")
		}
	} else {
		logrus.Debug("Docker container '%s' already running\n", dockerContainerName)
	}
}

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}
