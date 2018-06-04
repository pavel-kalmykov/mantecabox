package dao

import (
	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
)

func UserDaoFactory(engine string) UserDao {
	var userDao UserDao
	switch engine {
	case "postgres":
		userDao = UserPgDao{}
	default:
		logrus.Info(engine, "engine is not yet implemented")
		return nil
	}
	return userDao
}

func FileDaoFactory(engine string) FileDao {
	var fileDao FileDao
	switch engine {
	case "postgres":
		fileDao = FilePgDao{}
	default:
		log.Info(engine, "engine is not yet implemented")
		return nil
	}
	return fileDao
}

func LoginAttemptFactory(engine string) LoginAttempDao {
	switch engine {
	case "postgres":
		return LoginAttemptPgDao{}
	default:
		return nil
	}
}
