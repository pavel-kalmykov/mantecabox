package main

import (
	"fmt"
	_ "github.com/lib/pq"
	"testing"
	"database/sql"
	"time"
	_ "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert"
)

const (
	DBUSER     = "sds"
	DBPASSWORD = "sds"
	DBNAME     = "sds"
)

type UserInfo = struct {
	uid        int
	username   string
	department string
	created    time.Time
}

func TestDatabaseConnection(t *testing.T) {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		DBUSER, DBPASSWORD, DBNAME)
	db, err := sql.Open("postgres", dbinfo)
	assert.Nil(t, err)
	defer db.Close()

	// Inserting values
	var lastInsertId int
	err = db.QueryRow("INSERT INTO userinfo(username,departname,created) VALUES($1,$2,$3) RETURNING uid;", "astaxie", "研发部门", "2018-03-07").Scan(&lastInsertId)
	assert.Nil(t, err)
	assert.NotNil(t, lastInsertId)

	// Updating values
	stmt, err := db.Prepare("UPDATE userinfo SET username=$1 WHERE uid=$2")
	assert.Nil(t, err)
	res, err := stmt.Exec("astaxieupdate", lastInsertId)
	assert.Nil(t, err)
	affect, err := res.RowsAffected()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affect, "they should be equal")

	// Querying
	rows, err := db.Query("SELECT * FROM userinfo")
	assert.Nil(t, err)
	for rows.Next() {
		var userInfo UserInfo
		err = rows.Scan(&userInfo.uid, &userInfo.username, &userInfo.department, &userInfo.created)
		assert.Nil(t, err)
		assert.Equal(t, lastInsertId, userInfo.uid)
		assert.Equal(t, "astaxieupdate", userInfo.username)
		assert.Equal(t, "研发部门", userInfo.department)
		assert.Equal(t, "2018-03-07", userInfo.created.Format("2006-01-02"))
	}

	// Deleting
	stmt, err = db.Prepare("DELETE FROM userinfo WHERE uid=$1")
	assert.Nil(t, err)
	res, err = stmt.Exec(lastInsertId)
	assert.Nil(t, err)
	affect, err = res.RowsAffected()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affect, "they should be equal")
}
