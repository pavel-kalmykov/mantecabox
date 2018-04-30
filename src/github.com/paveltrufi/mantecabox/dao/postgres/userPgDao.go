package postgres

import (
	"database/sql"
	"errors"

	log "github.com/alexrudd/go-logger"
	"github.com/paveltrufi/mantecabox/models"
)

type UserPgDao struct {
}

func (dao UserPgDao) GetAll() ([]models.User, error) {
	users := make([]models.User, 0)
	db := GetPgDb()
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users WHERE deleted_at IS NULL")
	if err != nil {
		log.Info("Unable to execute UserPgDao.GetAll() query. Reason:", err)
		return nil, err
	}

	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.Username, &user.Password)
		if err != nil {
			log.Info("Unable to execute UserPgDao.GetAll() query. Reason:", err)
			return nil, err
		}
		users = append(users, user)
	}

	log.Debug("Queried", len(users), "users")
	return users, err
}

func (dao UserPgDao) GetByPk(username string) (models.User, error) {
	user := models.User{}
	db := GetPgDb()
	defer db.Close()

	err := db.QueryRow("SELECT * FROM users WHERE username = $1 AND deleted_at IS NULL", username).Scan(
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.Username, &user.Password)

	if err != nil {
		log.Debug("Unable to execute UserPgDao.GetByPk(username string) query. Reason:", err)
	} else {
		log.Debug("Retrieved user", user)
	}
	return user, err
}

func (dao UserPgDao) Create(user *models.User) (models.User, error) {
	db := GetPgDb()
	defer db.Close()

	var createdUser models.User
	err := db.QueryRow("INSERT INTO users(username,password) VALUES($1,$2) RETURNING *;",
		user.Username, user.Password).Scan(
		&createdUser.CreatedAt, &createdUser.UpdatedAt, &createdUser.DeletedAt, &createdUser.Username,
		&createdUser.Password)

	if err != nil {
		log.Info("Unable to execute UserPgDao.Create(user models.User) query. Reason:", err)
	} else {
		log.Debug("Created user", createdUser)
	}
	return createdUser, err
}

func (dao UserPgDao) Update(username string, user *models.User) (models.User, error) {
	db := GetPgDb()
	defer db.Close()

	var updatedUser models.User
	err := db.QueryRow("UPDATE users SET username=$1, password=$2 WHERE username=$3 RETURNING *",
		user.Username, user.Password, username).Scan(
		&updatedUser.CreatedAt, &updatedUser.UpdatedAt, &updatedUser.DeletedAt, &updatedUser.Username,
		&updatedUser.Password)

	if err != nil {
		log.Info("Unable to execute UserPgDao.Update(username string, user models.User) query. Reason:", err)
	} else {
		log.Debug("Updated user", updatedUser)
	}
	return updatedUser, err
}

func (dao UserPgDao) Delete(username string) error {
	db := GetPgDb()
	defer db.Close()

	result, err := db.Exec("DELETE FROM users WHERE username = $1", username)
	if err != nil {
		log.Info("Unable to execute UserPgDao.Delete(username string) query. Reason:", err)
	} else {
		var rowsAffected int64
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			log.Info("Some error occured during deleting:", err)
		} else {
			switch {
			case rowsAffected == 0:
				err = sql.ErrNoRows
			case rowsAffected > 1:
				err = errors.New("more than one deleted")
			}
			if err != nil {
				log.Debug("Unable to delete user with username \""+username+"\" correctly. Reason:", err)
			} else {
				log.Debug("User with username \"" + username + "\" successfully deleted")
			}
		}
	}
	return err
}
