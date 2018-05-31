package utilities

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"mantecabox/models"
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

func GetServerConfiguration() models.ServerConfig {
	configJsonPath := "./src/mantecabox/webservice/config.json"
	config := models.ServerConfig{}

	file, e := ioutil.ReadFile(configJsonPath)
	if e != nil {
		config.Port = "10443"
		config.Certificates.Cert = "cert.pem"
		config.Certificates.Key = "key.pem"
	}

	if config.Port == "" {
		config.Port = "10443"
	}

	if config.Certificates.Cert == "" {
		config.Certificates.Cert = "cert.pem"
	}

	if config.Certificates.Key == "" {
		config.Certificates.Key = "key.pem"
	}

	json.Unmarshal(file, &config)
	return config
}
