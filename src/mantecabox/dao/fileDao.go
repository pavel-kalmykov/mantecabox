package dao

import (
	"database/sql"
	"errors"
	"fmt"

	"mantecabox/models"
)

const (
	getAllFilesByOwnerQuery = `SELECT DISTINCT ON (name, owner) *
FROM (SELECT *
      FROM files f
        JOIN users u ON f.owner = u.email
      WHERE f.deleted_at IS NULL AND u.deleted_at IS NULL AND u.email = $1
      ORDER BY f.updated_at DESC) as T;`
	getFileVersionsByNameAndOwner = `SELECT
  f.*,
  u.*
FROM files f
  JOIN users u ON f.owner = u.email
WHERE f.deleted_at IS NULL AND u.deleted_at IS NULL AND f.name = $1 AND f.owner = $2`
	getLastVersionFileByNameAndOwnerQuery = `SELECT DISTINCT ON (name) *
FROM (SELECT *
      FROM files f
        JOIN users u ON f.owner = u.email
      WHERE f.deleted_at IS NULL AND u.deleted_at IS NULL AND f.name = $1 AND f.owner = $2
      ORDER BY f.updated_at DESC) as T;`
	getFileByVersionQuery = `SELECT f.*, u.* FROM files f JOIN users u on f.owner = u.email WHERE f.id = $1`
	insertFileQuery       = `INSERT INTO files (name, owner) VALUES ($1, $2) RETURNING *;`
	deleteFileQuery       = "UPDATE files SET deleted_at = NOW() WHERE name = $1 AND owner = $2"
)

type (
	FileDao interface {
		GetAllByOwner(user *models.User) ([]models.File, error)
		GetVersionsByNameAndOwner(filename string, user *models.User) ([]models.File, error)
		GetLastVersionFileByNameAndOwner(filename string, user *models.User) (models.File, error)
		GetFileByVersion(id int64) (models.File, error)
		Create(f *models.File) (models.File, error)
		Delete(filename string, user *models.User) error
	}

	FilePgDao struct {
	}
)

func (dao FilePgDao) GetAllByOwner(user *models.User) ([]models.File, error) {
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		files := make([]models.File, 0)
		rows, err := db.Query(getAllFilesByOwnerQuery, user.Email)
		if err != nil {
			daoLog.Info("Unable to execute FilePgDao.GetAllByOwner(user *models.User) query. Reason:", err)
			return nil, err
		}

		for rows.Next() {
			var file models.File
			err := scanFileRowWithUser(rows, &file)
			if err != nil {
				daoLog.Info("Unable to execute FilePgDao.GetAllByOwner(user *models.User) query. Reason:", err)
				return nil, err
			}
			files = append(files, file)
		}

		daoLog.Debug("Queried", len(files), "files")
		return files, err
	})
	return res.([]models.File), err
}

func (dao FilePgDao) GetVersionsByNameAndOwner(filename string, user *models.User) ([]models.File, error) {
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		files := make([]models.File, 0)
		rows, err := db.Query(getFileVersionsByNameAndOwner, filename, user.Email)
		if err != nil {
			daoLog.Info("Unable to execute FilePgDao.GetVersionsByNameAndOwner(filename string, user *models.User) query. Reason:", err)
			return nil, err
		}

		for rows.Next() {
			var file models.File
			err := scanFileRowWithUser(rows, &file)
			if err != nil {
				daoLog.Info("Unable to execute FilePgDao.GetVersionsByNameAndOwner(filename string, user *models.User) query. Reason:", err)
				return nil, err
			}
			files = append(files, file)
		}

		daoLog.Debug("Queried", len(files), "files")
		return files, err
	})
	return res.([]models.File), err
}

func (dao FilePgDao) GetLastVersionFileByNameAndOwner(filename string, user *models.User) (models.File, error) {
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		file := models.File{}
		row := db.QueryRow(getLastVersionFileByNameAndOwnerQuery, filename, user.Email)
		err := scanFileRowWithUser(row, &file)
		if err != nil {
			daoLog.Debug("Unable to execute FilePgDao.GetLastVersionFileByNameAndOwner(filename string, user *models.User) query. Reason:", err)
		} else {
			daoLog.Debug("Retrieved file ", file)
		}
		return file, err
	})
	return res.(models.File), err
}

func (dao FilePgDao) GetFileByVersion(id int64) (models.File, error) {
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		file := models.File{}
		row := db.QueryRow(getFileByVersionQuery, id)
		err := scanFileRowWithUser(row, &file)
		if err != nil {
			daoLog.Debug("Unable to execute FilePgDao.GetFileByVersion(id int64) query. Reason:", err)
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
				if rowsAffected == 0 {
					err = errors.New("not found")
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
		&file.Owner.Email,
		&file.PermissionsStr)
	return err
}

func scanFileRowWithUser(scanner polimorphicScanner, file *models.File) error {
	err := scanner.Scan(&file.Id,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.DeletedAt,
		&file.Name,
		&file.Owner.Email,
		&file.PermissionsStr,
		// user
		&file.Owner.CreatedAt,
		&file.Owner.UpdatedAt,
		&file.Owner.DeletedAt,
		&file.Owner.Email,
		&file.Owner.Password,
		&file.Owner.TwoFactorAuth,
		&file.Owner.TwoFactorTime)
	return err
}
