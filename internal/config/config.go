package config

import (
	"errors"
	"fmt"
	"os"
)

var errEnvVarNotFound error = errors.New("environment variable not found")

const (
	DB_CONNECTION_STRING_ENV     = "DB_CONNECTION_STRING"
	JWT_SECRET_ENV               = "JWT_SECRET"
	PORT_ENV                     = "PORT"
	MINIO_ADDRESS_ENV            = "MINIO_ADDRESS"
	MINIO_ACCESS_KEY_ENV         = "MINIO_ACCESS_KEY"
	MINIO_SECRET_KEY_ENV         = "MINIO_SECRET_KEY"
	MINIO_BUCKET_ENV             = "MINIO_BUCKET"
	RABBIT_CONNECTION_STRING_ENV = "RABBIT_CONNECTION_STRING"
)

type AppConfig struct {
	Port                   string
	DBConnectionString     string
	JWTSecret              string
	MinioAddress           string
	MinioAccessKey         string
	MinioSecretKey         string
	MinioBucketName        string
	RabbitConnectionString string
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

	minioAddress, ok := os.LookupEnv(MINIO_ADDRESS_ENV)
	if !ok {
		return AppConfig{}, fmt.Errorf("%w: %s", errEnvVarNotFound, MINIO_ADDRESS_ENV)
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

	rabbitConnStr, ok := os.LookupEnv(RABBIT_CONNECTION_STRING_ENV)
	if !ok {
		return AppConfig{}, fmt.Errorf("%w: %s", errEnvVarNotFound, RABBIT_CONNECTION_STRING_ENV)
	}

	return AppConfig{
		DBConnectionString:     connStr,
		JWTSecret:              jwtSecret,
		Port:                   port,
		MinioAddress:           minioAddress,
		MinioAccessKey:         minioAccessKey,
		MinioSecretKey:         minioSecretKey,
		MinioBucketName:        minioBucket,
		RabbitConnectionString: rabbitConnStr,
	}, nil
}
