package database

import (
	_ "github.com/lib/pq"
	"github.com/paveltrufi/mantecabox/utilities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

type UserInfo = struct {
	uid        int
	username   string
	department string
	created    time.Time
}

// TestDatabaseConnection es un test que prueba la conexión con la base de datos de Docker para comprobar su
// correcto funcionamiento
func TestDatabaseConnection(t *testing.T) {
	os.Setenv("MANTECABOX_CONFIG_FILE", "configuration.test.json")
	utilities.StartDockerPostgresDb()
	db, err := GetDbReadingConfig()
	require.NoError(t, err)
	defer db.Close()

	// Inserting values
	var lastInsertId int
	err = db.QueryRow("INSERT INTO userinfo(username,departname,created) VALUES($1,$2,$3) RETURNING uid;", "astaxie", "研发部门", "2018-03-07").Scan(&lastInsertId)
	require.NoError(t, err)
	assert.NotNil(t, lastInsertId)

	// Updating values
	stmt, err := db.Prepare("UPDATE userinfo SET username=$1 WHERE uid=$2")
	require.NoError(t, err)
	res, err := stmt.Exec("astaxieupdate", lastInsertId)
	require.NoError(t, err)
	affect, err := res.RowsAffected()
	require.NoError(t, err)
	assert.EqualValues(t, 1, affect)

	// Querying
	rows, err := db.Query("SELECT * FROM userinfo")
	require.NoError(t, err)
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
	require.NoError(t, err)
	res, err = stmt.Exec(lastInsertId)
	require.NoError(t, err)
	affect, err = res.RowsAffected()
	require.NoError(t, err)
	assert.EqualValues(t, 1, affect)
}
