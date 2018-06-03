package services

import (
	"mantecabox/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFileService(t *testing.T) {
	type args struct {
		configuration *models.Configuration
	}
	testCases := []struct {
		name string
		args args
		want FileService
	}{
		{
			name: "When passing the configuration, return the service",
			args: args{configuration: &models.Configuration{AesKey: "0123456789ABCDEF"}},
			want: FileServiceImpl{},
		},
		{
			name: "When passing no configuration, return nil",
			args: args{configuration: nil},
			want: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.IsType(t, testCase.want, NewFileService(testCase.args.configuration))
		})
	}
}
