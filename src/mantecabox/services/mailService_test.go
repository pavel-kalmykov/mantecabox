package services

import (
	"mantecabox/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSend2FAEmail(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "When the 2FA is sent correctly, return no error",
			test: func(t *testing.T) {
				require.NoError(t, testMailService.Send2FAEmail("mantecabox@gmail.com", "123456"))
			},
		},
		{
			name: "When the email does not exist, return an error",
			test: func(t *testing.T) {
				require.Error(t, testMailService.Send2FAEmail("unexistent@error.ko", "123456"))
			},
		},
		{
			name: "When the code is empty, return an error",
			test: func(t *testing.T) {
				require.Equal(t, Empty2FACodeError, testMailService.Send2FAEmail("mantecabox@gmail.com", ""))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func TestNewMailService(t *testing.T) {
	type args struct {
		configuration *models.Configuration
	}
	testCases := []struct {
		name string
		args args
		want MailService
	}{
		{
			name: "When passing the configuration, return the service",
			args: args{configuration: &models.Configuration{}},
			want: MailServiceImpl{},
		},
		{
			name: "When passing no configuration, return nil",
			args: args{configuration: nil},
			want: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.IsType(t, testCase.want, NewMailService(testCase.args.configuration))
		})
	}
}
