package postgres

import (
	"errors"

	"mantecabox/database"
	"mantecabox/models"

	"github.com/sirupsen/logrus"
)

type FilePgDao struct {
}

func (dao FilePgDao) GetAll(user *models.User) ([]models.File, error) {
	files := make([]models.File, 0)
	db, err := database.GetPgDb()
	defer db.Close()

	rows, err := db.Query(`SELECT
  f.*,
  u.*
FROM files f
  JOIN users u ON f.owner = u.email
WHERE f.deleted_at IS NULL AND u.deleted_at IS NULL AND u.email = $1`, user.Email)
	if err != nil {
		logrus.Info("Unable to execute FilePgDao.GetAll() query. Reason:", err)
		return nil, err
	}

	for rows.Next() {
		var file models.File
		var owner string
		err := rows.Scan(&file.Id, &file.CreatedAt, &file.UpdatedAt, &file.DeletedAt, &file.Name, &owner,
			&file.Owner.CreatedAt, &file.Owner.UpdatedAt, &file.Owner.DeletedAt, &file.Owner.Email, &file.Owner.Password, &file.Owner.TwoFactorAuth, &file.Owner.TwoFactorTime)
		if err != nil {
			logrus.Info("Unable to execute FilePgDao.GetAll() query. Reason:", err)
			return nil, err
		}
		files = append(files, file)
	}

	logrus.Debug("Queried", len(files), "files")
	return files, err
}

func (dao FilePgDao) GetByPk(filename string, user *models.User) (models.File, error) {
	file := models.File{}
	owner := ""
	db, err := database.GetPgDb()
	if err != nil {
		logrus.Fatal("Unable to connnect with database: " + err.Error())
	}
	defer db.Close()

	err = db.QueryRow(`SELECT
  f.*,
  u.*
FROM files f
  JOIN users u ON f.owner = u.email
WHERE f.deleted_at IS NULL AND u.deleted_at IS NULL AND f.name = $1 AND u.email = $2`, filename, user.Email).Scan(
		&file.Id, &file.CreatedAt, &file.UpdatedAt, &file.DeletedAt, &file.Name, &owner,
		&file.Owner.CreatedAt, &file.Owner.UpdatedAt, &file.Owner.DeletedAt, &file.Owner.Email, &file.Owner.Password, &file.Owner.TwoFactorAuth, &file.Owner.TwoFactorTime)

	if err != nil {
		logrus.Debug("Unable to execute FilePgDao.GetByPk(email string) query. Reason:", err)
	} else {
		logrus.Debug("Retrieved file ", file)
	}
	return file, err
}

func (dao FilePgDao) Create(file *models.File) (models.File, error) {
	db, err := database.GetPgDb()
	if err != nil {
		logrus.Fatal("Unable to connnect with database: " + err.Error())
	}
	defer db.Close()

	var createdFile models.File
	err = db.QueryRow(`INSERT INTO files (name, owner)
VALUES ($1, $2) RETURNING *;`, file.Name, file.Owner.Email,
	).Scan(&createdFile.Id, &createdFile.CreatedAt, &createdFile.UpdatedAt, &createdFile.DeletedAt,
		&createdFile.Name, &createdFile.Owner.Email)

	if err != nil {
		logrus.Info("Unable to execute FilePgDao.Create(file models.File) query. Reason:", err)
	} else {
		logrus.Debug("Created file: ", createdFile)
	}

	owner, err := UserPgDao{}.GetByPk(createdFile.Owner.Email)
	createdFile.Owner = owner

	return createdFile, err
}

func (dao FilePgDao) Update(id int64, file *models.File) (models.File, error) {
	db, err := database.GetPgDb()
	if err != nil {
		logrus.Fatal("Unable to connnect with database: " + err.Error())
	}
	defer db.Close()

	var updatedFile models.File
	err = db.QueryRow(`UPDATE files
SET name            = $1,
  owner             = $2
WHERE id = $3
RETURNING *`,
		file.Name, file.Owner.Email, id,
	).Scan(&updatedFile.Id, &updatedFile.CreatedAt, &updatedFile.UpdatedAt, &updatedFile.DeletedAt,
		&updatedFile.Name, &updatedFile.Owner.Email)

	if err != nil {
		logrus.Info("Unable to execute FilePgDao.Update(id int64, file models.File) query. Reason:", err)
	} else {
		logrus.Debug("Created file", updatedFile)
	}

	owner, err := UserPgDao{}.GetByPk(updatedFile.Owner.Email)
	updatedFile.Owner = owner

	return updatedFile, err
}

func (dao FilePgDao) Delete(id int64) error {
	db, err := database.GetPgDb()
	defer db.Close()

	result, err := db.Exec("DELETE FROM files WHERE id = $1", id)
	if err != nil {
		logrus.Info("Unable to execute FilePgDao.Delete(id int64) query. Reason:", err)
	} else {
		var rowsAffected int64
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			logrus.Info("Some error occured during deleting:", err)
		} else {
			switch {
			case rowsAffected == 0:
				err = errors.New("not found")
			case rowsAffected > 1:
				err = errors.New("more than one deleted")
			}
			if err != nil {
				logrus.Debug("Unable to delete file with id \""+string(id)+"\" correctly. Reason:", err)
			} else {
				logrus.Debug("File with id \"" + string(id) + "\" successfully deleted")
			}
		}
	}
	return err
}
