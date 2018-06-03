package utilities

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"

	"mantecabox/models"
)

func GetConfiguration() (models.Configuration, error) {
	config := models.Configuration{}
	filename, exists := os.LookupEnv("MANTECABOX_CONFIG_FILE")
	if !exists {
		if flag.Lookup("test.v") == nil {
			filename = "configuration.json"
		} else {
			filename = "configuration.test.json"
		}
	}
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
