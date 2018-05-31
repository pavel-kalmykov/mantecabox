package factory

import (
	"mantecabox/dao/interfaces"
	"mantecabox/dao/postgres"
)

func LoginAttemptFactory(engine string) interfaces.LoginAttempDao {
	switch engine {
	case "postgres":
		return postgres.LoginAttemptPgDao{}
	default:
		return nil
	}
}
