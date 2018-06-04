package dao

import (
	"database/sql"
	"errors"
	"fmt"

	"mantecabox/models"
)

const (
	getAllFilesQuery = `SELECT
  f.*,
  u.*
FROM files f
  JOIN users u ON f.owner = u.email
WHERE f.deleted_at IS NULL AND u.deleted_at IS NULL AND u.email = $1`
	getFileByNameAndOwner = `SELECT
  f.*,
  u.*
FROM files f
  JOIN users u ON f.owner = u.email
WHERE f.deleted_at IS NULL AND u.deleted_at IS NULL AND f.name = $1 AND u.email = $2`
	insertFileQuery = `INSERT INTO files (name, owner) VALUES ($1, $2) RETURNING *;`
	updateFileQuery = `UPDATE files SET name = $1, owner = $2 WHERE id = $3 RETURNING *`
	deleteFileQuery = "UPDATE files SET deleted_at = NOW() WHERE name = $1 AND owner = $2"
)

type (
	FileDao interface {
		GetAll(user *models.User) ([]models.File, error)
		GetByPk(filename string, user *models.User) (models.File, error)
		Create(f *models.File) (models.File, error)
		Update(id int64, f *models.File) (models.File, error)
		Delete(filename string, user *models.User) error
	}

	FilePgDao struct {
	}
)

func (dao FilePgDao) GetAll(user *models.User) ([]models.File, error) {
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		files := make([]models.File, 0)
		rows, err := db.Query(getAllFilesQuery, user.Email)
		if err != nil {
			daoLog.Info("Unable to execute FilePgDao.GetAll() query. Reason:", err)
			return nil, err
		}

		for rows.Next() {
			var file models.File
			err := scanFileRowWithUser(rows, &file)
			if err != nil {
				daoLog.Info("Unable to execute FilePgDao.GetAll() query. Reason:", err)
				return nil, err
			}
			files = append(files, file)
		}

		daoLog.Debug("Queried", len(files), "files")
		return files, err
	})
	return res.([]models.File), err
}

func (dao FilePgDao) GetByPk(filename string, user *models.User) (models.File, error) {
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		file := models.File{}
		row := db.QueryRow(getFileByNameAndOwner, filename, user.Email)
		err := scanFileRowWithUser(row, &file)
		if err != nil {
			daoLog.Debug("Unable to execute FilePgDao.GetByPk(email string) query. Reason:", err)
		} else {
			daoLog.Debug("Retrieved file ", file)
		}
		return file, err
	})
	return res.(models.File), err
}

func (dao FilePgDao) Create(file *models.File) (models.File, error) {
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		var createdFile models.File
		row := db.QueryRow(insertFileQuery, file.Name, file.Owner.Email)
		err := scanFileRow(row, &createdFile)
		if err != nil {
			daoLog.Info("Unable to execute FilePgDao.Create(file models.File) query. Reason:", err)
		} else {
			daoLog.Debug("Created file: ", createdFile)
		}
		owner, err := UserPgDao{}.GetByPk(createdFile.Owner.Email)
		createdFile.Owner = owner
		return createdFile, err
	})
	return res.(models.File), err
}

func (dao FilePgDao) Update(id int64, file *models.File) (models.File, error) {
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		var updatedFile models.File
		row := db.QueryRow(updateFileQuery,
			file.Name, file.Owner.Email, id,
		)
		err := scanFileRow(row, &updatedFile)
		if err != nil {
			daoLog.Info("Unable to execute FilePgDao.Update(id int64, file models.File) query. Reason:", err)
		} else {
			daoLog.Debug("Created file", updatedFile)
		}
		owner, err := UserPgDao{}.GetByPk(updatedFile.Owner.Email)
		updatedFile.Owner = owner
		return updatedFile, err
	})
	return res.(models.File), err
}

func (dao FilePgDao) Delete(filename string, user *models.User) error {
	_, err := withDb(func(db *sql.DB) (interface{}, error) {
		result, err := db.Exec(deleteFileQuery, filename, user.Email)
		if err != nil {
			daoLog.Info("Unable to execute FilePgDao.Delete(id int64) query. Reason:", err)
		} else {
			var rowsAffected int64
			rowsAffected, err = result.RowsAffected()
			if err != nil {
				daoLog.Info("Some error occured during deleting:", err)
			} else {
				switch {
				case rowsAffected == 0:
					err = errors.New("not found")
				case rowsAffected > 1:
					err = errors.New("more than one deleted")
				}
				if err != nil {
					daoLog.Debug(fmt.Sprintf(`Unable to delete %v's' file "%v". Reason %v`, user.Email, filename, err))
				} else {
					daoLog.Debug(fmt.Sprintf(`%v's' file "%v" successfully deleted.`, user.Email, filename))
				}
			}
		}
		return nil, err
	})
	return err
}

func scanFileRow(scanner polimorphicScanner, file *models.File) error {
	err := scanner.Scan(&file.Id,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.DeletedAt,
		&file.Name,
		&file.Owner.Email)
	return err
}

func scanFileRowWithUser(scanner polimorphicScanner, file *models.File) error {
	err := scanner.Scan(&file.Id,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.DeletedAt,
		&file.Name,
		&file.Owner.Email,
		&file.Owner.CreatedAt,
		&file.Owner.UpdatedAt,
		&file.Owner.DeletedAt,
		&file.Owner.Email,
		&file.Owner.Password,
		&file.Owner.TwoFactorAuth,
		&file.Owner.TwoFactorTime)
	return err
}
