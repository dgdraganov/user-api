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
	MINIO_ENDPOINT_ENV       = "MINIO_ENDPOINT"
	MINIO_ACCESS_KEY_ENV     = "MINIO_ACCESS_KEY"
	MINIO_SECRET_KEY_ENV     = "MINIO_SECRET_KEY"
	MINIO_BUCKET_ENV         = "MINIO_BUCKET"
)

type AppConfig struct {
	Port               string
	DBConnectionString string
	JWTSecret          string
	MinioEndpoint      string
	MinioAccessKey     string
	MinioSecretKey     string
	MinioBucketName    string
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

	minioEndpoint, ok := os.LookupEnv(MINIO_ENDPOINT_ENV)
	if !ok {
		return AppConfig{}, fmt.Errorf("%w: %s", errEnvVarNotFound, MINIO_ENDPOINT_ENV)
	}

	minioAccessKey, ok := os.LookupEnv(MINIO_ACCESS_KEY_ENV)
	if !ok {
		return AppConfig{}, fmt.Errorf("%w: %s", errEnvVarNotFound, MINIO_ACCESS_KEY_ENV)
	}

	minioSecretKey, ok := os.LookupEnv(MINIO_SECRET_KEY_ENV)
	if !ok {
		return AppConfig{}, fmt.Errorf("%w: %s", errEnvVarNotFound, MINIO_SECRET_KEY_ENV)
	}

	minioBucket, ok := os.LookupEnv(MINIO_BUCKET_ENV)
	if !ok {
		return AppConfig{}, fmt.Errorf("%w: %s", errEnvVarNotFound, MINIO_BUCKET_ENV)
	}

	return AppConfig{
		DBConnectionString: connStr,
		JWTSecret:          jwtSecret,
		Port:               port,
		MinioEndpoint:      minioEndpoint,
		MinioAccessKey:     minioAccessKey,
		MinioSecretKey:     minioSecretKey,
		MinioBucketName:    minioBucket,
	}, nil
}
