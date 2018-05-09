package factory

import (
	"testing"

	"github.com/paveltrufi/mantecabox/dao/interfaces"
	"github.com/paveltrufi/mantecabox/dao/postgres"
	"github.com/stretchr/testify/require"
)

func TestUserDaoFactory(t *testing.T) {
	type args struct {
		engine string
	}
	testCases := []struct {
		name string
		args args
		want interfaces.UserDao
	}{
		{
			`When asking for "postgres" DAO, return userPgDao instance`,
			args{engine: "postgres"},
			postgres.UserPgDao{},
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
				require.IsType(t, postgres.UserPgDao{}, UserDaoFactory(testCase.args.engine))
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
		want interfaces.FileDao
	}{
		{
			name: `When asking for "postgres" DAO, return userPgDao instance`,
			args: args{engine: "postgres"},
			want: postgres.FilePgDao{},
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
				require.IsType(t, postgres.FilePgDao{}, FileDaoFactory(testCase.args.engine))
			}
		})
	}
}
