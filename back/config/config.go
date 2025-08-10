package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GoEnv          string
	AuthSecret     string
	StorageType    string
	DataSourceName string
	Port           string
	LogFileName    string
	Values         map[string]value
}

type value struct {
	Value   string
	Default string
}

func LoadConfig() Config {
	godotenv.Load()

	cfg := Config{
		GoEnv:          "GOENV",
		Port:           "PORT",
		AuthSecret:     "AUTH_SECRET",
		StorageType:    "STORAGE_TYPE",
		DataSourceName: "DATA_SOURCE_NAME",
		LogFileName:    "LOG_FILE_NAME",
		Values:         make(map[string]value),
	}

	cfg.Add(cfg.GoEnv, value{Value: os.Getenv("GOENV"), Default: "dev"})
	cfg.Add(cfg.Port, value{Value: os.Getenv("PORT"), Default: "9090"})
	cfg.Add(cfg.AuthSecret, value{Value: os.Getenv("AUTH_SECRET"), Default: "totally secret"})
	cfg.Add(cfg.StorageType, value{Value: os.Getenv("STORAGE_TYPE"), Default: "sqlite"})
	cfg.Add(cfg.LogFileName, value{Value: os.Getenv("LOG_FILE_NAME"), Default: cfg.Get(cfg.GoEnv) + ".json"})
	cfg.Add(cfg.DataSourceName, value{Value: os.Getenv("DATA_SOURCE_NAME"), Default: cfg.Get(cfg.GoEnv) + ".sqlite"})

	return cfg
}

func (c Config) Get(name string) string {
	returnVal := ""
	val, ok := c.Values[name]
	if !ok {
		log.Fatal(fmt.Errorf("Field %s in config was not defined", name))
	}
	if val.Value != "" {
		returnVal = val.Value
	} else if val.Default != "" {
		returnVal = val.Default
	} else {
		log.Fatal(fmt.Errorf("Field %s in config is empty", name))
	}
	return returnVal
}

func (c *Config) Add(name string, value value) {
	c.Values[name] = value
}
