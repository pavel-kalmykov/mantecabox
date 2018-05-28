package factory

import (
	"mantecabox/dao/interfaces"
	"mantecabox/dao/postgres"

	log "github.com/alexrudd/go-logger"
)

func UserDaoFactory(engine string) interfaces.UserDao {
	var userDao interfaces.UserDao
	switch engine {
	case "postgres":
		userDao = postgres.UserPgDao{}
	default:
		log.Info(engine, "engine is not yet implemented")
		return nil
	}
	return userDao
}
