package postgres

import (
	"errors"

	log "github.com/alexrudd/go-logger"
	"github.com/paveltrufi/mantecabox/models"
)

type FilePgDao struct {
}

func (dao FilePgDao) GetAll() ([]models.File, error) {
	files := make([]models.File, 0)
	db := GetPgDb()
	defer db.Close()

	rows, err := db.Query(`SELECT
  f.*,
  u.*
FROM files f
  JOIN users u ON f.owner = u.username
WHERE f.deleted_at IS NULL AND u.deleted_at IS NULL `)
	if err != nil {
		log.Info("Unable to execute FilePgDao.GetAll() query. Reason:", err)
		return nil, err
	}

	for rows.Next() {
		var file models.File
		var owner string
		err := rows.Scan(&file.Id, &file.CreatedAt, &file.UpdatedAt, &file.DeletedAt, &file.Name, &owner, &file.Group,
			&file.UserReadable, &file.UserWritable, &file.UserExecutable,
			&file.GroupReadable, &file.GroupWritable, &file.GroupExecutable,
			&file.OtherReadable, &file.OtherWritable, &file.OtherExecutable,
			&file.PlatformCreation,
			&file.Owner.CreatedAt, &file.Owner.UpdatedAt, &file.Owner.DeletedAt, &file.Owner.Username, &file.Owner.Password)
		if err != nil {
			log.Info("Unable to execute FilePgDao.GetAll() query. Reason:", err)
			return nil, err
		}
		files = append(files, file)
	}

	log.Debug("Queried", len(files), "files")
	return files, err
}

func (dao FilePgDao) GetByPk(id int64) (models.File, error) {
	file := models.File{}
	owner := ""
	db := GetPgDb()
	defer db.Close()

	err := db.QueryRow(`SELECT
  f.*,
  u.*
FROM files f
  JOIN users u ON f.owner = u.username
WHERE f.deleted_at IS NULL AND u.deleted_at IS NULL AND f.id = $1`, id).Scan(
		&file.Id, &file.CreatedAt, &file.UpdatedAt, &file.DeletedAt, &file.Name, &owner, &file.Group,
		&file.UserReadable, &file.UserWritable, &file.UserExecutable,
		&file.GroupReadable, &file.GroupWritable, &file.GroupExecutable,
		&file.OtherReadable, &file.OtherWritable, &file.OtherExecutable,
		&file.PlatformCreation,
		&file.Owner.CreatedAt, &file.Owner.UpdatedAt, &file.Owner.DeletedAt, &file.Owner.Username, &file.Owner.Password)

	if err != nil {
		log.Debug("Unable to execute FilePgDao.GetByPk(username string) query. Reason:", err)
	} else {
		log.Debug("Retrieved file ", file)
	}
	return file, err
}

func (dao FilePgDao) Create(file *models.File) (models.File, error) {
	db := GetPgDb()
	defer db.Close()

	var createdFile models.File
	err := db.QueryRow(`INSERT INTO files (name, owner, "group", user_readable, user_writable, user_executable, group_readable, group_writable, group_executable, other_readable, other_writable, other_executable, platform_creation)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING *;`, file.Name, file.Owner.Username, file.Group,
		file.UserReadable, file.UserWritable, file.UserExecutable,
		file.GroupReadable, file.GroupWritable, file.GroupExecutable,
		file.OtherReadable, file.OtherWritable, file.OtherExecutable,
		file.PlatformCreation,
	).Scan(&createdFile.Id, &createdFile.CreatedAt, &createdFile.UpdatedAt, &createdFile.DeletedAt,
		&createdFile.Name, &createdFile.Owner.Username, &createdFile.Group,
		&createdFile.UserReadable, &createdFile.UserWritable, &createdFile.UserExecutable,
		&createdFile.GroupReadable, &createdFile.GroupWritable, &createdFile.GroupExecutable,
		&createdFile.OtherReadable, &createdFile.OtherWritable, &createdFile.OtherExecutable,
		&createdFile.PlatformCreation)

	if err != nil {
		log.Info("Unable to execute FilePgDao.Create(file models.File) query. Reason:", err)
	} else {
		log.Debug("Created file", createdFile)
	}

	owner, err := UserPgDao{}.GetByPk(createdFile.Owner.Username)
	createdFile.Owner = owner

	return createdFile, err
}

func Update(id int64, file *models.File) (models.File, error) {
	db := GetPgDb()
	defer db.Close()

	var updatedFile models.File
	err := db.QueryRow(`UPDATE files
SET name            = $1,
  owner             = $2,
  "group"           = $3,
  user_readable     = $4,
  user_writable     = $5,
  user_executable   = $6,
  group_readable    = $7,
  group_writable    = $8,
  group_executable  = $9,
  other_readable    = $10,
  other_writable    = $11,
  other_executable  = $12,
  platform_creation = $13
WHERE id = $14
RETURNING *`,
		file.Name, file.Owner.Username, file.Group,
		file.UserReadable, file.UserWritable, file.UserExecutable,
		file.GroupReadable, file.GroupWritable, file.GroupExecutable,
		file.OtherReadable, file.OtherWritable, file.OtherExecutable,
		file.PlatformCreation, id,
	).Scan(&updatedFile.Id, &updatedFile.CreatedAt, &updatedFile.UpdatedAt, &updatedFile.DeletedAt,
		&updatedFile.Name, &updatedFile.Owner.Username, &updatedFile.Group,
		&updatedFile.UserReadable, &updatedFile.UserWritable, &updatedFile.UserExecutable,
		&updatedFile.GroupReadable, &updatedFile.GroupWritable, &updatedFile.GroupExecutable,
		&updatedFile.OtherReadable, &updatedFile.OtherWritable, &updatedFile.OtherExecutable,
		&updatedFile.PlatformCreation)

	if err != nil {
		log.Info("Unable to execute FilePgDao.Update(id int64, file models.File) query. Reason:", err)
	} else {
		log.Debug("Created file", updatedFile)
	}

	owner, err := UserPgDao{}.GetByPk(updatedFile.Owner.Username)
	updatedFile.Owner = owner

	return updatedFile, err
}

func Delete(id int64) error {
	db := GetPgDb()
	defer db.Close()

	result, err := db.Exec("DELETE FROM files WHERE id = $1", id)
	if err != nil {
		log.Info("Unable to execute FilePgDao.Delete(id int64) query. Reason:", err)
	} else {
		var rowsAffected int64
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			log.Info("Some error occured during deleting:", err)
		} else {
			switch {
			case rowsAffected == 0:
				err = errors.New("not found")
			case rowsAffected > 1:
				err = errors.New("more than one deleted")
			}
			if err != nil {
				log.Debug("Unable to delete file with id \""+string(id)+"\" correctly. Reason:", err)
			} else {
				log.Debug("File with id \"" + string(id) + "\" successfully deleted")
			}
		}
	}
	return err
}
