package factory

import (
	"mantecabox/dao/interfaces"
	"mantecabox/dao/postgres"

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
