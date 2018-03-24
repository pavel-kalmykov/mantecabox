package factory

import (
	log "github.com/alexrudd/go-logger"
	"github.com/paveltrufi/mantecabox/dao/interfaces"
	"github.com/paveltrufi/mantecabox/dao/postgres"
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
