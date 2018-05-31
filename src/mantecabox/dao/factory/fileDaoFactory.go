package factory

import (
	"mantecabox/dao/interfaces"
	"mantecabox/dao/postgres"

	log "github.com/alexrudd/go-logger"
)

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
