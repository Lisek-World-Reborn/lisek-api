package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

type ApiConfiguration struct {
	Version int    `yaml:"version"`
	Port    int    `yaml:"port"`
	Dsn     string `yaml:"dsn"`
	Secret  string `yaml:"secret"`
	Redis   struct {
		Address  string `yaml:"address"`
		Password string `yaml:"password"`
	} `yaml:"redis"`
}

var LoadedConfiguration ApiConfiguration

func IsConfigurationExists(filepath string) bool {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return false
	}

	return true
}

func GetConfiguration(filepath string) (*ApiConfiguration, error) {
	if !IsConfigurationExists(filepath) {
		return nil, nil
	}

	cfgBytes, err := ioutil.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	var cfg ApiConfiguration

	err = yaml.Unmarshal(cfgBytes, &cfg)

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func GenerateDefaultConfiguration(filepath string) error {
	cfg := ApiConfiguration{
		Version: 1,
		Port:    8080,
		Dsn:     "host=localhost user=lisek password=lisek dbname=lisek port=5432 sslmode=disable TimeZone=Europe/Kiev",
		Secret:  "secret",
		Redis: struct {
			Address  string `yaml:"address"`
			Password string `yaml:"password"`
		}{
			Address:  "localhost:6379",
			Password: "",
		},
	}

	cfgBytes, err := yaml.Marshal(cfg)

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath, cfgBytes, 0644)

	if err != nil {
		return err
	}

	return nil
}
