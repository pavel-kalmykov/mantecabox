package dao

import (
	"mantecabox/logs"

)

func UserDaoFactory(engine string) UserDao {
	logs.DaoLog.Debug("UserDaoFactory")
	var userDao UserDao
	switch engine {
	case "postgres":
		userDao = UserPgDao{}
	default:
		logs.DaoLog.Info(engine, "engine is not yet implemented")
		return nil
	}
	return userDao
}

func FileDaoFactory(engine string) FileDao {
	logs.DaoLog.Debug("FileDaoFactory")
	var fileDao FileDao
	switch engine {
	case "postgres":
		fileDao = FilePgDao{}
	default:
		logs.DaoLog.Info(engine, "engine is not yet implemented")
		return nil
	}
	return fileDao
}

func LoginAttemptFactory(engine string) LoginAttempDao {
	logs.DaoLog.Debug("LoginAttemptFactory")
	switch engine {
	case "postgres":
		return LoginAttemptPgDao{}
	default:
		return nil
	}
}
