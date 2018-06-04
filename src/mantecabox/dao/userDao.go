package dao

import (
	"database/sql"
	"errors"

	"mantecabox/logs"
	"mantecabox/models"

	"github.com/PeteProgrammer/go-automapper"
	"github.com/sirupsen/logrus"
)

const (
	getAllUsersQuery = "SELECT * FROM users WHERE deleted_at IS NULL"
	getUserByPkQuery = "SELECT * FROM users WHERE deleted_at IS NULL AND email = $1"
	insertUserQuery  = "INSERT INTO users(email,password) VALUES($1,$2) RETURNING *;"
	updateUserQuery  = "UPDATE users SET email=$1, password=$2, two_factor_auth=$3 WHERE email=$4 RETURNING *"
	deleteUserQuery  = "UPDATE users SET deleted_at = NOW() WHERE email = $1"
)

var daoLog = logrus.WithFields(logrus.Fields{"package": "postgres"})

type (
	UserPgDao struct {
	}

	UserDao interface {
		GetAll() ([]models.User, error)
		GetByPk(username string) (models.User, error)
		Create(u *models.User) (models.User, error)
		Update(username string, u *models.User) (models.User, error)
		Delete(username string) error
	}
)

func (dao UserPgDao) GetAll() ([]models.User, error) {
	logs.DaoLog.Debug("GetAll")
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		users := make([]models.User, 0)
		rows, err := db.Query(getAllUsersQuery)
		if err != nil {
			logs.DaoLog.Infof("Unable to execute UserPgDao.GetAll() query. Reason: %v", err)
			return nil, err
		}

		for rows.Next() {
			var user models.User
			err := scanUserRow(rows, &user)
			if err != nil {
				logs.DaoLog.Infof("Unable to execute UserPgDao.GetAll() query. Reason: %v", err)
				return nil, err
			}
			users = append(users, user)
		}

		logs.DaoLog.Info("Queried ", len(users), " users")
		return users, err
	})
	return res.([]models.User), err
}

func (dao UserPgDao) GetByPk(email string) (models.User, error) {
	logs.DaoLog.Debug("GetByPk")
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		user := models.User{}
		row := db.QueryRow(getUserByPkQuery, email)
		err := scanUserRow(row, &user)
		if err != nil {
			logs.DaoLog.Infof("Unable to execute UserPgDao.GetVersionsByNameAndOwner(email string) query. Reason: %v", err)
		} else {
			var dto models.UserDto
			automapper.Map(user, &dto)
			logs.DaoLog.Infof("Retrieved user %v", dto)
		}
		return user, err
	})
	return res.(models.User), err
}

func (dao UserPgDao) Create(user *models.User) (models.User, error) {
	logs.DaoLog.Debug("Create")
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		var createdUser models.User
		row := db.QueryRow(insertUserQuery,
			user.Email, user.Password)
		err := scanUserRow(row, &createdUser)
		if err != nil {
			logs.DaoLog.Infof("Unable to execute UserPgDao.Create(user models.User) query. Reason: %v", err)
		} else {
			logs.DaoLog.Infof("Created user %v", createdUser)
		}
		return createdUser, err
	})
	return res.(models.User), err
}

func (dao UserPgDao) Update(email string, user *models.User) (models.User, error) {
	logs.DaoLog.Debug("Update")
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		var updatedUser models.User
		row := db.QueryRow(updateUserQuery,
			user.Email, user.Password, user.TwoFactorAuth, email)
		err := scanUserRow(row, &updatedUser)
		if err != nil {
			logs.DaoLog.Infof("Unable to execute UserPgDao.Update(email string, user models.User) query. Reason: %v", err)
		} else {
			logs.DaoLog.Infof("Updated user %v", updatedUser)
		}
		return updatedUser, err
	})
	return res.(models.User), err
}

func (dao UserPgDao) Delete(email string) error {
	logs.DaoLog.Debug("Delete")
	_, err := withDb(func(db *sql.DB) (interface{}, error) {
		result, err := db.Exec(deleteUserQuery, email)
		if err != nil {
			logs.DaoLog.Infof("Unable to execute UserPgDao.Delete(email string) query. Reason: %v", err)
		} else {
			var rowsAffected int64
			rowsAffected, err = result.RowsAffected()
			if err != nil {
				logs.DaoLog.Info("Some error occured during deleting:", err)
			} else {
				switch {
				case rowsAffected == 0:
					err = sql.ErrNoRows
				case rowsAffected > 1:
					err = errors.New("more than one deleted")
				}
				if err != nil {
					logs.DaoLog.Info("Unable to delete user with email \""+email+"\" correctly. Reason:", err)
				} else {
					logs.DaoLog.Info("User with email \"" + email + "\" successfully deleted")
				}
			}
		}
		return nil, err
	})
	return err
}

type polimorphicScanner interface {
	Scan(dest ...interface{}) error
}

func scanUserRow(scanner polimorphicScanner, user *models.User) error {
	logs.DaoLog.Debug("scanUserRow")
	err := scanner.Scan(
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.Email,
		&user.Password,
		&user.TwoFactorAuth,
		&user.TwoFactorTime)
	return err
}
