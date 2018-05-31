package factory

import (
	"mantecabox/dao/interfaces"
	"mantecabox/dao/postgres"

	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoginAttemptFactory(t *testing.T) {
	type args struct {
		engine string
	}
	tests := []struct {
		name string
		args args
		want interfaces.LoginAttempDao
	}{
		{
			`When asking for "postgres" DAO, return userPgDao instance`,
			args{engine: "postgres"},
			postgres.LoginAttemptPgDao{},
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
