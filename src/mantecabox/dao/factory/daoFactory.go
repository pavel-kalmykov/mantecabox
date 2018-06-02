package factory

import (
	"mantecabox/dao/interfaces"
	"mantecabox/dao/postgres"

	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
)

func UserDaoFactory(engine string) interfaces.UserDao {
	var userDao interfaces.UserDao
	switch engine {
	case "postgres":
		userDao = postgres.UserPgDao{}
	default:
		logrus.Info(engine, "engine is not yet implemented")
		return nil
	}
	return userDao
}

func FileDaoFactory(engine string) interfaces.FileDao {
	var userDao interfaces.FileDao
	switch engine {
	case "postgres":
		userDao = postgres.FilePgDao{}
	default:
		log.Info(engine, "engine is not yet implemented")
		return nil
	}
	return userDao
}

func LoginAttemptFactory(engine string) interfaces.LoginAttempDao {
	switch engine {
	case "postgres":
		return postgres.LoginAttemptPgDao{}
	default:
		return nil
	}
}
