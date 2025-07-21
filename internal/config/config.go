package config

import (
	"errors"
	"fmt"
	"os"
)

var errEnvVarNotFound error = errors.New("environment variable not found")

const (
	DB_CONNECTION_STRING_ENV = "DB_CONNECTION_STRING"
	JWT_SECRET_ENV           = "JWT_SECRET"
	PORT_ENV                 = "PORT"
)

type AppConfig struct {
	Port               string
	DBConnectionString string
	JWTSecret          string
}

func NewAppConfig() (AppConfig, error) {

	connStr, ok := os.LookupEnv(DB_CONNECTION_STRING_ENV)
	if !ok {
		return AppConfig{}, fmt.Errorf("%w: %s", errEnvVarNotFound, DB_CONNECTION_STRING_ENV)
	}

	jwtSecret, ok := os.LookupEnv(JWT_SECRET_ENV)
	if !ok {
		return AppConfig{}, fmt.Errorf("%w: %s", errEnvVarNotFound, JWT_SECRET_ENV)
	}

	port, ok := os.LookupEnv(PORT_ENV)
	if !ok {
		return AppConfig{}, fmt.Errorf("%w: %s", errEnvVarNotFound, PORT_ENV)
	}

	return AppConfig{
		DBConnectionString: connStr,
		JWTSecret:          jwtSecret,
		Port:               port,
	}, nil
}
