package utilities

import (
	"mantecabox/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseManagerImpl_RunMigrations(t *testing.T) {
	type fields struct {
		database *models.Database
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "When database configuration is correct, run the migrations properly",
			fields: fields{
				database: &models.Database{
					Engine:   "postgres",
					Host:     "localhost",
					Port:     5432,
					User:     "sds",
					Password: "sds",
					Name:     "sds_test",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			databaseManager := DatabaseManagerImpl{
				database: tt.fields.database,
			}
			err := databaseManager.RunMigrations()
			require.NoError(t, err)
		})
	}
}
