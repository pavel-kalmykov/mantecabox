package utilities

import (
	"encoding/json"
	"github.com/paveltrufi/mantecabox/models"
	"os"
)

func GetConfiguration() (models.Configuration, error) {
	config := models.Configuration{}
	filename, exists := os.LookupEnv("MANTECABOX_CONFIG_FILE")
	if !exists {
		filename = "configuration.json"
	}
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil

}
