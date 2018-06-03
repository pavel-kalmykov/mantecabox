package dao

import (
	"os"
	"testing"

	"mantecabox/utilities"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	configuration, err := utilities.GetConfiguration()
	if err != nil {
		logrus.Fatal(err)
	}
	databaseManager := utilities.NewDatabaseManager(&configuration.Database)
	err = databaseManager.StartDockerPostgresDb()
	if err != nil {
		logrus.Fatal(err)
	}
	err = databaseManager.RunMigrations()
	if err != nil {
		logrus.Fatal(err)
	}
	os.Exit(m.Run())
}

func TestUserDaoFactory(t *testing.T) {
	type args struct {
		engine string
	}
	testCases := []struct {
		name string
		args args
		want UserDao
	}{
		{
			`When asking for "postgres" DAO, return userPgDao instance`,
			args{engine: "postgres"},
			UserPgDao{},
		},
		{
			`When asking for "mysql" DAO, return nil`,
			args{engine: "mysql"},
			nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			switch testCase.want {
			case nil:
				require.Nil(t, UserDaoFactory(testCase.args.engine))
			default:
				require.IsType(t, UserPgDao{}, UserDaoFactory(testCase.args.engine))
			}
		})
	}
}

func TestFileDaoFactory(t *testing.T) {
	type args struct {
		engine string
	}
	testCases := []struct {
		name string
		args args
		want FileDao
	}{
		{
			name: `When asking for "postgres" DAO, return userPgDao instance`,
			args: args{engine: "postgres"},
			want: FilePgDao{},
		},
		{
			name: `When asking for "mysql" DAO, return nil`,
			args: args{engine: "mysql"},
			want: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			switch testCase.want {
			case nil:
				require.Nil(t, FileDaoFactory(testCase.args.engine))
			default:
				require.IsType(t, FilePgDao{}, FileDaoFactory(testCase.args.engine))
			}
		})
	}
}

func TestLoginAttemptFactory(t *testing.T) {
	type args struct {
		engine string
	}
	tests := []struct {
		name string
		args args
		want LoginAttempDao
	}{
		{
			`When asking for "postgres" DAO, return userPgDao instance`,
			args{engine: "postgres"},
			LoginAttemptPgDao{},
		},
		{
			`When asking for "mysql" DAO, return nil`,
			args{engine: "mysql"},
			nil,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.want, LoginAttemptFactory(testCase.args.engine))
		})
	}
}
