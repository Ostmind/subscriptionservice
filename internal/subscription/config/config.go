package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Srv      ServerConfig   `yaml:"server"`
	DB       DatabaseConfig `yaml:"db"`
	LogLevel string         `yaml:"env"`
}

type ServerConfig struct {
	Host                  string        `yaml:"host"`
	Port                  int           `yaml:"port"`
	ServerReadTimeout     time.Duration `yaml:"read-timeout"`
	ServerWriteTimeout    time.Duration `yaml:"write-timeout"`
	ServerShutdownTimeout time.Duration `yaml:"shutdown-timeout"`
	MigrationPath         string        `yaml:"migration"`
}

type DatabaseConfig struct {
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	DBName     string `yaml:"db-name"`
	DBUser     string `yaml:"db-user"`
	DBPassword string `yaml:"db-password"`
	DBSSLMode  string `yaml:"db-ssl-mode"`
}

func MustNew() *AppConfig {
	configPath, err := fetchConfigPath()
	if err != nil {
		log.Fatalf("error fetching config file: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("error reading config file: %v", err)
	}

	var cfg AppConfig

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("error unmarshaling YAML: %v", err)
	}

	errs := cfg.Validate()
	if errs != nil {
		log.Fatalf("err validating config: %s", errs.Error())
	}

	return &cfg
}

func (cfg *AppConfig) Validate() (result error) {
	if cfg.Srv.Host == "" {
		result = errors.Join(result, ErrNoServerHost)
	}

	if cfg.Srv.Port == 0 {
		result = errors.Join(result, ErrNoServerPort)
	}

	if cfg.DB.Host == "" {
		result = errors.Join(result, ErrNoDBHost)
	}

	if cfg.DB.Port == "" {
		result = errors.Join(result, ErrNoDBPort)
	}

	if cfg.DB.DBName == "" {
		result = errors.Join(result, ErrNoDBName)
	}

	if cfg.DB.DBUser == "" {
		result = errors.Join(result, ErrNoDBUser)
	}

	if cfg.DB.DBPassword == "" {
		result = errors.Join(result, ErrNoDBPassword)
	}

	return result
}

func fetchConfigPath() (string, error) {
	var path string

	flag.StringVar(&path, "config", "", "path to config file")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")

		if path == "" {
			return "", errors.New("config path is required")
		}
	}

	return path, nil
}

func GetConnStr(host, port, user, password, dbName, sslMode string) string {
	return fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslMode)
}
