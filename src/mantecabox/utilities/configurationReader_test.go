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
	"engine": "postgres",
	"server": "localhost",
	"port": "5432",
	"user": "sds",
	"password": "sds",
	"database": "sds"
}`,
			models.Configuration{
				Engine:   "postgres",
				Server:   "localhost",
				Port:     "5432",
				User:     "sds",
				Password: "sds",
				Database: "sds",
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
	"engine": "postgres",
	"server": "localhost",
	"port": "5432"
}`,
			models.Configuration{
				Engine: "postgres",
				Server: "localhost",
				Port:   "5432",
			},
			false,
		},
		{
			"When the config file is not properly formatted, it must return an empty object and an error",
			`{
	"engine": "postgres",
	"server": "localhost",
	"port": "5432", <-- that ending comma -and this comment- is a JSON format error
}`,
			models.Configuration{},
			true,
		},
	}

	for _, tt := range testCases {
		// Test preparation
		os.Remove("configuration.test.json")
		if tt.config != "" {
			ioutil.WriteFile("configuration.test.json", []byte(tt.config), os.ModePerm)
		}

		// Test execution
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConfiguration()
			if tt.wantErr {
				require.Error(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}
