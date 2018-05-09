package postgres

import (
	"testing"

	"github.com/aodin/date"
	"github.com/lib/pq"
	"github.com/paveltrufi/mantecabox/models"
	"github.com/stretchr/testify/require"
)

const (
	testUsersInsertQuery = `INSERT INTO users (username, password)
VALUES ('testuser1', 'testpassword1'),
  ('testuser2', 'testpassword2');`

	testFileInsertQuery = `INSERT INTO files (name, owner) VALUES ('testfile1a', 'testuser1') RETURNING id;`
)

func TestFilePgDao_GetAll(t *testing.T) {
	testCases := []struct {
		name        string
		insertQuery string
		want        []models.File
	}{
		{
			"When the files table has some files, retrieve all them",
			testUsersInsertQuery + `INSERT INTO files (name, owner)
VALUES ('testfile1a', 'testuser1'),
  ('testfile1b', 'testuser1'),
  ('testfile2a', 'testuser2'),
  ('testfile2b', 'testuser2');`,
			[]models.File{
				{Name: "testfile1a", Owner: models.User{Credentials: models.Credentials{Username: "testuser1", Password: "testpassword1"}}, UserReadable: true, UserWritable: true, GroupReadable: true},
				{Name: "testfile1b", Owner: models.User{Credentials: models.Credentials{Username: "testuser1", Password: "testpassword1"}}, UserReadable: true, UserWritable: true, GroupReadable: true},
				{Name: "testfile2a", Owner: models.User{Credentials: models.Credentials{Username: "testuser2", Password: "testpassword2"}}, UserReadable: true, UserWritable: true, GroupReadable: true},
				{Name: "testfile2b", Owner: models.User{Credentials: models.Credentials{Username: "testuser2", Password: "testpassword2"}}, UserReadable: true, UserWritable: true, GroupReadable: true},
			},
		},
		{
			"When the files table is empty, retrieve an empty set",
			``,
			[]models.File{},
		},
		{
			"When the files table has some deleted users, don't retrieve them",
			testUsersInsertQuery + `INSERT INTO files (deleted_at, name, owner)
VALUES (NULL, 'testfile1a', 'testuser1'),
  (NOW(), 'testfile1b', 'testuser1'),
  (NULL, 'testfile2a', 'testuser2'),
  (NOW(), 'testfile2b', 'testuser2');`,
			[]models.File{
				{Name: "testfile1a", Owner: models.User{Credentials: models.Credentials{Username: "testuser1", Password: "testpassword1"}}, UserReadable: true, UserWritable: true, GroupReadable: true},
				{Name: "testfile2a", Owner: models.User{Credentials: models.Credentials{Username: "testuser2", Password: "testpassword2"}}, UserReadable: true, UserWritable: true, GroupReadable: true},
			},
		},
	}

	db := getDb(t)
	defer db.Close()

	for _, testCase := range testCases {
		cleanAndPopulateDb(db, testCase.insertQuery, t)

		t.Run(testCase.name, func(t *testing.T) {
			dao := FilePgDao{}
			got, err := dao.GetAll()
			require.NoError(t, err)

			// We ignore the timestamps as we don't need to get them compared
			// But we check they are valid (they were created today)
			// Then, we also skip the ID
			for k, v := range got {
				createdAtDate := date.FromTime(v.CreatedAt.Time)
				updatedAtDate := date.FromTime(v.UpdatedAt.Time)
				require.True(t, createdAtDate.Within(date.SingleDay(createdAtDate)))
				require.True(t, updatedAtDate.Within(date.SingleDay(updatedAtDate)))
				got[k].CreatedAt = pq.NullTime{}
				got[k].UpdatedAt = pq.NullTime{}

				// same for owners
				createdAtDate = date.FromTime(v.Owner.CreatedAt.Time)
				updatedAtDate = date.FromTime(v.Owner.UpdatedAt.Time)
				require.True(t, createdAtDate.Within(date.SingleDay(createdAtDate)))
				require.True(t, updatedAtDate.Within(date.SingleDay(updatedAtDate)))
				got[k].Owner.CreatedAt = pq.NullTime{}
				got[k].Owner.UpdatedAt = pq.NullTime{}

				got[k].Id = 0
			}
			require.Equal(t, testCase.want, got)
		})
	}
}

func TestFilePgDao_GetByPk(t *testing.T) {
	type args struct {
		id int64
	}
	testCases := []struct {
		name        string
		insertQuery string
		args        args
		want        models.File
		wantErr     bool
	}{
		{
			"When you ask for an existent file, retrieve it",
			testFileInsertQuery,
			args{},
			models.File{Name: "testfile1a", Owner: models.User{Credentials: models.Credentials{Username: "testuser1", Password: "testpassword1"}}, UserReadable: true, UserWritable: true, GroupReadable: true},
			false,
		},
		{
			"When you ask for an non-existent file, return an empty file and an error",
			``,
			args{id: -1},
			models.File{},
			true,
		},
		{
			"When you ask for a deleted file, return an empty file and an error",
			`INSERT INTO files (deleted_at, name, owner) VALUES (NOW(), 'testfile1a', 'testuser1') RETURNING id;`,
			args{},
			models.File{},
			true,
		},
	}

	db := getDb(t)
	defer db.Close()

	for _, testCase := range testCases {
		cleanAndPopulateDb(db, testUsersInsertQuery, t)
		if testCase.insertQuery != "" {
			db.QueryRow(testCase.insertQuery).Scan(&testCase.want.Id)
			if testCase.wantErr {
				testCase.want.Id = 0
			} else {
				testCase.args.id = testCase.want.Id
			}
		}
		t.Run(testCase.name, func(t *testing.T) {
			dao := FilePgDao{}
			got, err := dao.GetByPk(testCase.args.id)
			requireFileEqualCheckingErrors(t, testCase.wantErr, err, testCase.want, got)
		})
	}
}

func TestFilePgDao_Create(t *testing.T) {
	file := models.File{Name: "testfile", Owner: models.User{Credentials: models.Credentials{Username: "testuser1", Password: "testpassword1"}},
		UserReadable: true, UserWritable: true, GroupReadable: true, GroupWritable: true}
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

	db := GetPgDb()
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

func TestUpdate(t *testing.T) {
	updatedFile := models.File{Name: "updatedfile", Owner: models.User{Credentials: models.Credentials{Username: "testuser1", Password: "testpassword1"}}}
	updatedFileWithoutFilename := updatedFile
	updatedFileWithoutFilename.Name = ""
	updatedFileWithoutOwner := updatedFile
	updatedFileWithoutOwner.Owner = models.User{}
	type args struct {
		id   int64
		file *models.File
	}
	testCases := []struct {
		name        string
		insertQuery string
		args        args
		want        models.File
		correctId   bool
		wantErr     bool
	}{
		{
			"When you update an already inserted file, return the file updated",
			testFileInsertQuery,
			args{file: &updatedFile},
			updatedFile,
			true,
			false,
		},
		{
			"When you update a non-existent file, return an empty file and an error",
			testFileInsertQuery,
			args{file: &updatedFile},
			models.File{},
			false,
			true,
		},
		{
			"When you update a file without filename, return an empty file and an error",
			testFileInsertQuery,
			args{file: &updatedFileWithoutFilename},
			models.File{},
			true,
			true,
		},
		{
			"When you update a file without owner, return an empty file and an error",
			testFileInsertQuery,
			args{file: &updatedFileWithoutOwner},
			models.File{},
			true,
			true,
		},
	}

	db := GetPgDb()
	defer db.Close()

	for _, testCase := range testCases {
		cleanAndPopulateDb(db, testUsersInsertQuery, t)
		if testCase.insertQuery != "" {
			db.QueryRow(testCase.insertQuery).Scan(&testCase.want.Id)
			if testCase.correctId {
				testCase.args.id = testCase.want.Id
			}
			if testCase.wantErr {
				testCase.want.Id = 0
			}
		}
		t.Run(testCase.name, func(t *testing.T) {
			dao := FilePgDao{}
			got, err := dao.Update(testCase.args.id, testCase.args.file)
			requireFileEqualCheckingErrors(t, testCase.wantErr, err, testCase.want, got)
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		id int64
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
			args{},
			false,
		},
		{
			"When you delete a non-existent file, return an error",
			testFileInsertQuery,
			args{id: -1},
			true,
		},
	}

	db := getDb(t)
	defer db.Close()

	for _, testCase := range testCases {
		cleanAndPopulateDb(db, testUsersInsertQuery, t)
		if testCase.insertQuery != "" {
			db.QueryRow(testCase.insertQuery).Scan(&testCase.args.id)
			if testCase.wantErr {
				testCase.args.id = 0
			}
		}
		t.Run(testCase.name, func(t *testing.T) {
			dao := FilePgDao{}
			err := dao.Delete(testCase.args.id)
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
		actual.CreatedAt = pq.NullTime{}
		actual.UpdatedAt = pq.NullTime{}

		// same for owners
		createdAtDate = date.FromTime(actual.Owner.CreatedAt.Time)
		updatedAtDate = date.FromTime(actual.Owner.UpdatedAt.Time)
		require.True(t, createdAtDate.Within(date.SingleDay(createdAtDate)))
		require.True(t, updatedAtDate.Within(date.SingleDay(updatedAtDate)))
		actual.Owner.CreatedAt = pq.NullTime{}
		actual.Owner.UpdatedAt = pq.NullTime{}
	}
	require.Equal(t, expected, actual)
}
