package dao

import (
	"database/sql"
	"errors"
	"fmt"

	"mantecabox/logs"
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
	setGDriveIdQuery      = `UPDATE files SET gdrive_id = $1 WHERE id = $2`
	deleteFileQuery       = "UPDATE files SET deleted_at = NOW() WHERE name = $1 AND owner = $2"
)

type (
	FileDao interface {
		GetAllByOwner(user *models.User) ([]models.File, error)
		GetVersionsByNameAndOwner(filename string, user *models.User) ([]models.File, error)
		GetLastVersionFileByNameAndOwner(filename string, user *models.User) (models.File, error)
		GetFileByVersion(id int64) (models.File, error)
		Create(f *models.File) (models.File, error)
		SetGdriveId(id int64, gdriveId string) error
		Delete(filename string, user *models.User) error
	}

	FilePgDao struct {
	}
)

func (dao FilePgDao) GetAllByOwner(user *models.User) ([]models.File, error) {
	logs.DaoLog.Debug("GetAllByOwner")
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		files := make([]models.File, 0)
		rows, err := db.Query(getAllFilesByOwnerQuery, user.Email)
		if err != nil {
			logs.DaoLog.Errorf("Unable to execute FilePgDao.GetAllByOwner(user *models.User) query. Reason: %v", err)
			return nil, err
		}

		for rows.Next() {
			var file models.File
			err := scanFileRowWithUser(rows, &file)
			if err != nil {
				logs.DaoLog.Errorf("Unable to execute FilePgDao.GetAllByOwner(user *models.User) query. Reason: %v", err)
				return nil, err
			}
			files = append(files, file)
		}

		logs.DaoLog.Info("Queried ", len(files), " files")
		return files, err
	})
	return res.([]models.File), err
}

func (dao FilePgDao) GetVersionsByNameAndOwner(filename string, user *models.User) ([]models.File, error) {
	logs.DaoLog.Debug("GetVersionsByNameAndOwner")
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		files := make([]models.File, 0)
		rows, err := db.Query(getFileVersionsByNameAndOwner, filename, user.Email)
		if err != nil {
			logs.DaoLog.Errorf("Unable to execute FilePgDao.GetVersionsByNameAndOwner(filename string, user *models.User) query. Reason: %v", err)
			return nil, err
		}

		for rows.Next() {
			var file models.File
			err := scanFileRowWithUser(rows, &file)
			if err != nil {
				logs.DaoLog.Errorf("Unable to execute FilePgDao.GetVersionsByNameAndOwner(filename string, user *models.User) query. Reason: %v", err)
				return nil, err
			}
			files = append(files, file)
		}

		logs.DaoLog.Info("Queried ", len(files), " files")
		return files, err
	})
	return res.([]models.File), err
}

func (dao FilePgDao) GetLastVersionFileByNameAndOwner(filename string, user *models.User) (models.File, error) {
	logs.DaoLog.Debug("GetLastVersionFileByNameAndOwner")
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		file := models.File{}
		row := db.QueryRow(getLastVersionFileByNameAndOwnerQuery, filename, user.Email)
		err := scanFileRowWithUser(row, &file)
		if err != nil {
			logs.DaoLog.Infof("Unable to execute FilePgDao.GetLastVersionFileByNameAndOwner(filename string, user *models.User) query. Reason: %v", err)
		} else {
			logs.DaoLog.Infof("Retrieved file %v", models.FileToDto(file))
		}
		return file, err
	})
	return res.(models.File), err
}

func (dao FilePgDao) GetFileByVersion(id int64) (models.File, error) {
	logs.DaoLog.Debug("GetFileByVersion")
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		file := models.File{}
		row := db.QueryRow(getFileByVersionQuery, id)
		err := scanFileRowWithUser(row, &file)
		if err != nil {
			logs.DaoLog.Infof("Unable to execute FilePgDao.GetFileByVersion(id int64) query. Reason: %v", err)
		} else {
			logs.DaoLog.Infof("Retrieved file %v", models.FileToDto(file))
		}
		return file, err
	})
	return res.(models.File), err
}

func (dao FilePgDao) Create(file *models.File) (models.File, error) {
	logs.DaoLog.Debug("Create")
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		var createdFile models.File
		row := db.QueryRow(insertFileQuery, file.Name, file.Owner.Email)
		err := scanFileRow(row, &createdFile)
		if err != nil {
			logs.DaoLog.Errorf("Unable to execute FilePgDao.Create(file models.File) query. Reason: %v", err)
		} else {
			logs.DaoLog.Infof("Created file: %v", createdFile)
		}
		owner, err := UserPgDao{}.GetByPk(createdFile.Owner.Email)
		createdFile.Owner = owner
		return createdFile, err
	})
	return res.(models.File), err
}

func (dao FilePgDao) SetGdriveId(id int64, gdriveId string) error {
	logs.DaoLog.Debug("SetGdriveId")
	_, err := withDb(func(db *sql.DB) (interface{}, error) {
		result, err := db.Exec(setGDriveIdQuery, gdriveId, id)
		if err != nil {
			logs.DaoLog.Errorf("Unable to execute FilePgDao.Delete(id int64) query. Reason: %v", err)
		} else {
			var rowsAffected int64
			rowsAffected, err = result.RowsAffected()
			if err != nil {
				logs.DaoLog.Errorf("Some error occured during setting gdrive id: %v", err)
			} else {
				if rowsAffected == 0 {
					err = errors.New("not found")
				}
				if err != nil {
					logs.DaoLog.Info(fmt.Sprintf(`Unable to set %v's file gdrive id "%v". Reason %v`, id, gdriveId, err))
				} else {
					logs.DaoLog.Info(fmt.Sprintf(`%v's' file gdrive id "%v" successfully set.`, id, gdriveId))
				}
			}
		}
		return nil, err
	})
	return err
}

func (dao FilePgDao) Delete(filename string, user *models.User) error {
	logs.DaoLog.Debug("Delete")
	_, err := withDb(func(db *sql.DB) (interface{}, error) {
		result, err := db.Exec(deleteFileQuery, filename, user.Email)
		if err != nil {
			logs.DaoLog.Error("Unable to execute FilePgDao.Delete(id int64) query. Reason:", err)
		} else {
			var rowsAffected int64
			rowsAffected, err = result.RowsAffected()
			if err != nil {
				logs.DaoLog.Error("Some error occured during deleting:", err)
			} else {
				if rowsAffected == 0 {
					err = errors.New("not found")
				}
				if err != nil {
					logs.DaoLog.Info(fmt.Sprintf(`Unable to delete %v's' file "%v". Reason %v`, user.Email, filename, err))
				} else {
					logs.DaoLog.Info(fmt.Sprintf(`%v's' file "%v" successfully deleted.`, user.Email, filename))
				}
			}
		}
		return nil, err
	})
	return err
}

func scanFileRow(scanner polimorphicScanner, file *models.File) error {
	logs.DaoLog.Debug("scanFileRow")
	err := scanner.Scan(&file.Id,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.DeletedAt,
		&file.Name,
		&file.Owner.Email,
		&file.PermissionsStr,
		&file.GdriveID)
	return err
}

func scanFileRowWithUser(scanner polimorphicScanner, file *models.File) error {
	logs.DaoLog.Debug("scanFileRowWithUser")
	err := scanner.Scan(&file.Id,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.DeletedAt,
		&file.Name,
		&file.Owner.Email,
		&file.PermissionsStr,
		&file.GdriveID,
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
