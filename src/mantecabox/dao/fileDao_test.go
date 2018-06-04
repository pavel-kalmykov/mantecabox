package dao

import (
	"testing"

	"mantecabox/models"
	"mantecabox/utilities"

	"github.com/aodin/date"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v3"
)

const (
	testUsersInsertQuery = `INSERT INTO users (email, password)
VALUES ('testuser1', 'testpassword1'),
  ('testuser2', 'testpassword2');`

	testFileInsertQuery = `INSERT INTO files (name, owner) VALUES ('testfile1a', 'testuser1') RETURNING id;`
)

func TestFilePgDao_GetAllByOwner(t *testing.T) {
	type args struct {
		user models.User
	}
	testCases := []struct {
		name        string
		insertQuery string
		want        []models.File
		args        args
	}{
		{
			"When the files table has some files, retrieve all them",
			testUsersInsertQuery + `INSERT INTO files (name, owner)
VALUES ('testfile1a', 'testuser1'),
  ('testfile1b', 'testuser1'),
  ('testfile2a', 'testuser2'),
  ('testfile2b', 'testuser2');`,
			[]models.File{
				{Name: "testfile1a", Owner: models.User{Credentials: models.Credentials{Email: "testuser1", Password: "testpassword1"}}, PermissionsStr: "rw-r--r--"},
				{Name: "testfile1b", Owner: models.User{Credentials: models.Credentials{Email: "testuser1", Password: "testpassword1"}}, PermissionsStr: "rw-r--r--"},
			},
			args{models.User{Credentials: models.Credentials{Email: "testuser1", Password: "testpassword1"}}},
		},
		{
			"When the files table is empty, retrieve an empty set",
			``,
			[]models.File{},
			args{models.User{Credentials: models.Credentials{Email: "testuser1", Password: "testpassword1"}}},
		},
		{
			"When the files table has some deleted users, don't retrieve them",
			testUsersInsertQuery + `INSERT INTO files (deleted_at, name, owner)
VALUES (NULL, 'testfile1a', 'testuser1'),
  (NOW(), 'testfile1b', 'testuser1'),
  (NULL, 'testfile2a', 'testuser2'),
  (NOW(), 'testfile2b', 'testuser2');`,
			[]models.File{
				{Name: "testfile1a", Owner: models.User{Credentials: models.Credentials{Email: "testuser1", Password: "testpassword1"}}, PermissionsStr: "rw-r--r--"},
			},
			args{models.User{Credentials: models.Credentials{Email: "testuser1", Password: "testpassword1"}}},
		},
	}

	db := getDb(t)
	defer db.Close()

	for _, testCase := range testCases {
		cleanAndPopulateDb(db, testCase.insertQuery, t)

		t.Run(testCase.name, func(t *testing.T) {
			dao := FilePgDao{}
			got, err := dao.GetAllByOwner(&testCase.args.user)
			require.NoError(t, err)

			// We ignore the timestamps as we don't need to get them compared
			// But we check they are valid (they were created today)
			// Then, we also skip the ID
			for k, v := range got {
				createdAtDate := date.FromTime(v.CreatedAt.Time)
				updatedAtDate := date.FromTime(v.UpdatedAt.Time)
				require.True(t, createdAtDate.Within(date.SingleDay(createdAtDate)))
				require.True(t, updatedAtDate.Within(date.SingleDay(updatedAtDate)))
				got[k].CreatedAt = null.Time{}
				got[k].UpdatedAt = null.Time{}

				// same for owners
				createdAtDate = date.FromTime(v.Owner.CreatedAt.Time)
				updatedAtDate = date.FromTime(v.Owner.UpdatedAt.Time)
				require.True(t, createdAtDate.Within(date.SingleDay(createdAtDate)))
				require.True(t, updatedAtDate.Within(date.SingleDay(updatedAtDate)))
				got[k].Owner.CreatedAt = null.Time{}
				got[k].Owner.UpdatedAt = null.Time{}

				got[k].Id = 0
			}
			require.Equal(t, testCase.want, got)
		})
	}
}

func TestFilePgDao_GetByPk(t *testing.T) {
	type args struct {
		filename string
		user     *models.User
	}
	testCases := []struct {
		name        string
		insertQuery string
		args        args
		want        []models.File
	}{
		{
			"When you ask for an existent file, retrieve it",
			testFileInsertQuery,
			args{"testfile1a", &models.User{Credentials: models.Credentials{Email: "testuser1"}}},
			[]models.File{
				{Name: "testfile1a", Owner: models.User{Credentials: models.Credentials{Email: "testuser1", Password: "testpassword1"}}, PermissionsStr: "rw-r--r--"},
			},
		},
		{
			"When you ask for an non-existent file, return an empty file and an error",
			``,
			args{"noexiste", &models.User{Credentials: models.Credentials{Email: "noexiste"}}},
			[]models.File{},
		},
		{
			"When you ask for a deleted file, return an empty file and an error",
			`INSERT INTO files (deleted_at, name, owner) VALUES (NOW(), 'testfile1a', 'testuser1') RETURNING id;`,
			args{"testfile1a", &models.User{Credentials: models.Credentials{Email: "testuser1"}}},
			[]models.File{},
		},
	}

	db := getDb(t)
	defer db.Close()

	for _, testCase := range testCases {
		cleanAndPopulateDb(db, testUsersInsertQuery+testCase.insertQuery, t)
		t.Run(testCase.name, func(t *testing.T) {
			dao := FilePgDao{}
			got, err := dao.GetVersionsByNameAndOwner(testCase.args.filename, testCase.args.user)
			require.NoError(t, err)
			for k, v := range got {
				createdAtDate := date.FromTime(v.CreatedAt.Time)
				updatedAtDate := date.FromTime(v.UpdatedAt.Time)
				require.True(t, createdAtDate.Within(date.SingleDay(createdAtDate)))
				require.True(t, updatedAtDate.Within(date.SingleDay(updatedAtDate)))
				got[k].CreatedAt = null.Time{}
				got[k].UpdatedAt = null.Time{}

				// same for owners
				createdAtDate = date.FromTime(v.Owner.CreatedAt.Time)
				updatedAtDate = date.FromTime(v.Owner.UpdatedAt.Time)
				require.True(t, createdAtDate.Within(date.SingleDay(createdAtDate)))
				require.True(t, updatedAtDate.Within(date.SingleDay(updatedAtDate)))
				got[k].Owner.CreatedAt = null.Time{}
				got[k].Owner.UpdatedAt = null.Time{}

				got[k].Id = 0
			}
			require.Equal(t, testCase.want, got)
		})
	}
}

func TestFilePgDao_Create(t *testing.T) {
	file := models.File{Name: "testfile", Owner: models.User{Credentials: models.Credentials{Email: "testuser1", Password: "testpassword1"}}, PermissionsStr: "rw-r--r--"}
	fileWithoutName := file
	fileWithoutName.Name = ""
	fileWithoutOwner := file
	fileWithoutOwner.Owner = models.User{}

	type args struct {
		file *models.File
	}
	testCases := []struct {
		name        string
		insertQuery string
		args        args
		want        models.File
		wantErr     bool
	}{
		{
			"When you create a new file, it gets inserted",
			testUsersInsertQuery,
			args{file: &file},
			file,
			false,
		},
		{
			"When you create a new file without filename, return an empty file and an error",
			testUsersInsertQuery,
			args{file: &fileWithoutName},
			models.File{},
			true,
		},
		{
			"When you create a new file without owner, return an empty file and an error",
			testUsersInsertQuery,
			args{file: &fileWithoutOwner},
			models.File{},
			true,
		},
	}

	db, err := utilities.GetPgDb()
	if err != nil {
		logrus.Fatal("Unable to connnect with database: " + err.Error())
	}
	defer db.Close()

	for _, testCase := range testCases {
		cleanAndPopulateDb(db, testCase.insertQuery, t)
		t.Run(testCase.name, func(t *testing.T) {
			dao := FilePgDao{}
			got, err := dao.Create(testCase.args.file)
			got.Id = 0 // Ignoramos el ID porque no podemos saber cuál es de antemano
			requireFileEqualCheckingErrors(t, testCase.wantErr, err, testCase.want, got)
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		filename string
		user     *models.User
	}
	testCases := []struct {
		name        string
		insertQuery string
		args        args
		wantErr     bool
	}{
		{
			"When you delete an inserted file, return no error",
			testFileInsertQuery,
			args{
				filename: "testfile1a",
				user:     &models.User{Credentials: models.Credentials{Email: "testuser1"}},
			},
			false,
		},
		{
			"When you delete a non-existent file, return an error",
			testFileInsertQuery,
			args{
				filename: "nonexistent",
				user:     &models.User{},
			},
			true,
		},
	}

	db := getDb(t)
	defer db.Close()

	for _, testCase := range testCases {
		cleanAndPopulateDb(db, testUsersInsertQuery, t)
		if testCase.insertQuery != "" {
			db.QueryRow(testCase.insertQuery)
		}
		t.Run(testCase.name, func(t *testing.T) {
			dao := FilePgDao{}
			err := dao.Delete(testCase.args.filename, testCase.args.user)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func requireFileEqualCheckingErrors(t *testing.T, wantErr bool, err error, expected models.File, actual models.File) {
	if wantErr {
		require.Error(t, err)
		require.Equal(t, models.File{}, actual)
	} else {
		require.NoError(t, err)
		// We ignore the timestamps as we don't need to get them compared
		// But we check they are valid (they were created recently)
		createdAtDate := date.FromTime(actual.CreatedAt.Time)
		updatedAtDate := date.FromTime(actual.UpdatedAt.Time)
		require.True(t, createdAtDate.Within(date.SingleDay(createdAtDate)))
		require.True(t, updatedAtDate.Within(date.SingleDay(updatedAtDate)))

		// same for owners
		createdAtDate = date.FromTime(actual.Owner.CreatedAt.Time)
		updatedAtDate = date.FromTime(actual.Owner.UpdatedAt.Time)
		require.True(t, createdAtDate.Within(date.SingleDay(createdAtDate)))
		require.True(t, updatedAtDate.Within(date.SingleDay(updatedAtDate)))
	}
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Owner.Email, actual.Owner.Email)
	require.Equal(t, expected.Owner.Password, actual.Owner.Password)
}
