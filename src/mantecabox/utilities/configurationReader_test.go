package utilities

import (
	"mantecabox/models"

	"github.com/stretchr/testify/require"

	"io/ioutil"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("MANTECABOX_CONFIG_FILE", "configuration.test.json")
	code := m.Run()
	os.Remove("configuration.test.json")
	os.Exit(code)
}

func TestGetConfiguration(t *testing.T) {
	testCases := []struct {
		name    string
		config  string
		want    models.Configuration
		wantErr bool
	}{
		{
			"When the config file is OK, it must return the proper configuration",
			`{
  "aes_key": "6368616e676520746869732070617373",
  "token_timeout": "1h",
  "blocked_login_time_limit": "5m",
  "max_unsuccessful_attempts": 3,
  "files_path": "files/",
  "database": {
    "engine": "postgres",
    "host": "localhost",
    "port": 5432,
    "user": "sds",
    "password": "sds",
    "name": "sds"
  },
  "server": {
    "host": "localhost",
    "port": 10443,
    "cert": "cert.pem",
    "key": "key.pem"
  },
  "mail": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "mantecabox@gmail.com",
    "password": "ElChiringuito"
  }
}
`,
			models.Configuration{
				AesKey:                  "6368616e676520746869732070617373",
				TokenTimeout:            "1h",
				BlockedLoginTimeLimit:   "5m",
				MaxUnsuccessfulAttempts: 3,
				FilesPath:               "files/",
				Database: models.Database{
					Engine:   "postgres",
					Host:     "localhost",
					Port:     5432,
					User:     "sds",
					Password: "sds",
					Name:     "sds",
				},
				Server: models.Server{
					Host: "localhost",
					Port: 10443,
					Cert: "cert.pem",
					Key:  "key.pem",
				},
				Mail: models.Mail{
					Host:     "smtp.gmail.com",
					Port:     587,
					Username: "mantecabox@gmail.com",
					Password: "ElChiringuito",
				},
			},
			false,
		},
		{
			"When there is no file, it must return an empty object and an error",
			"",
			models.Configuration{},
			true,
		},
		{
			"When the config file is incomplete, it must return an incomplete object",
			`{
  "aes_key": "6368616e676520746869732070617373",
  "token_timeout": "1h",
  "blocked_login_time_limit": "5m",
  "max_unsuccessful_attempts": 3,
  "files_path": "files/"
}
`,
			models.Configuration{
				AesKey:                  "6368616e676520746869732070617373",
				TokenTimeout:            "1h",
				BlockedLoginTimeLimit:   "5m",
				MaxUnsuccessfulAttempts: 3,
				FilesPath:               "files/",
				Database:                models.Database{},
				Server:                  models.Server{},
				Mail:                    models.Mail{},
			},
			false,
		},
		{
			"When the config file is not properly formatted, it must return an empty object and an error",
			`{
	"engine": "postgres",
	"server": "localhost",
	"port": "5432", // <-- that ending comma -and this comment- is a JSON format error
}`,
			models.Configuration{},
			true,
		},
	}

	for _, tt := range testCases {
		os.Remove("configuration.test.json")
		if tt.config != "" {
			ioutil.WriteFile("configuration.test.json", []byte(tt.config), os.ModePerm)
		}

		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConfiguration()
			if tt.wantErr {
				require.Error(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}
