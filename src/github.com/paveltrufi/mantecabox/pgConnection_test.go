package main

import (
	"fmt"
	_ "github.com/lib/pq"
	"testing"
	"database/sql"
	"time"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"log"
	"strings"
)

const (
	DBUSER              = "sds"
	DBPASSWORD          = "sds"
	DBNAME              = "sds"
	dockerContainerName = "sds-postgres"
)

type UserInfo = struct {
	uid        int
	username   string
	department string
	created    time.Time
}

// startDockerPostgresDb ejecuta un comando para comprobar si el contenedor de la base de datos está en ejecución o no.
// Este comando devolverá "true\n" o "false\n", así que comprobamos que si no devuelve true lo iniciemos (esto lo
// sabremos si al ejecutarse el comando éste ha devuelto el nombre del mismo seguido de un salto de línea).
func startDockerPostgresDb() {
	command := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", dockerContainerName)
	output, err := command.Output()
	checkErr(err)
	// usamos hasPrefix para no tener que controlar los saltos de línea
	if !(strings.HasPrefix(string(output), "true")) {
		output, err := exec.Command("docker", "container", "start", dockerContainerName).Output()
		checkErr(err)
		if strings.HasPrefix(string(output), dockerContainerName) {
			log.Printf("Docker container '%s' started\n", dockerContainerName)
		} else {
			panic("Unable to start Postgre's docker container!")
		}
	} else {
		log.Printf("Docker container '%s' already running\n", dockerContainerName)
	}
}

func TestDatabaseConnection(t *testing.T) {
	startDockerPostgresDb()
	databaseInfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		DBUSER, DBPASSWORD, DBNAME)
	db, err := sql.Open("postgres", databaseInfo)
	checkErr(err)
	defer db.Close()

	// Inserting values
	var lastInsertId int
	err = db.QueryRow("INSERT INTO userinfo(username,departname,created) VALUES($1,$2,$3) RETURNING uid;", "astaxie", "研发部门", "2018-03-07").Scan(&lastInsertId)
	checkErr(err)
	assert.NotNil(t, lastInsertId)

	// Updating values
	stmt, err := db.Prepare("UPDATE userinfo SET username=$1 WHERE uid=$2")
	checkErr(err)
	res, err := stmt.Exec("astaxieupdate", lastInsertId)
	checkErr(err)
	affect, err := res.RowsAffected()
	checkErr(err)
	assert.EqualValues(t, 1, affect)

	// Querying
	rows, err := db.Query("SELECT * FROM userinfo")
	checkErr(err)
	for rows.Next() {
		var userInfo UserInfo
		err = rows.Scan(&userInfo.uid, &userInfo.username, &userInfo.department, &userInfo.created)
		assert.Nil(t, err)
		assert.EqualValues(t, lastInsertId, userInfo.uid)
		assert.EqualValues(t, "astaxieupdate", userInfo.username)
		assert.EqualValues(t, "研发部门", userInfo.department)
		assert.EqualValues(t, "2018-03-07", userInfo.created.Format("2006-01-02"))
	}

	// Deleting
	stmt, err = db.Prepare("DELETE FROM userinfo WHERE uid=$1")
	checkErr(err)
	res, err = stmt.Exec(lastInsertId)
	checkErr(err)
	affect, err = res.RowsAffected()
	checkErr(err)
	assert.EqualValues(t, 1, affect)
}

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}
