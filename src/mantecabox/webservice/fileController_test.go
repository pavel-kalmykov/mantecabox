package webservice

import (
	"mantecabox/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFileController(t *testing.T) {
	type args struct {
		configuration *models.Configuration
	}
	testCases := []struct {
		name string
		args args
		want FileController
	}{
		{
			name: "When passing the configuration, return the service",
			args: args{configuration: &models.Configuration{AesKey: "0123456789ABCDEF"}},
			want: FileControllerImpl{},
		},
		{
			name: "When passing no configuration, return nil",
			args: args{configuration: nil},
			want: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.IsType(t, testCase.want, NewFileController(testCase.args.configuration))
		})
	}
}
