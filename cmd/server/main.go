package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dgdraganov/user-api/internal/config"
	"github.com/dgdraganov/user-api/internal/core"
	"github.com/dgdraganov/user-api/internal/db"
	"github.com/dgdraganov/user-api/internal/http/handler"
	"github.com/dgdraganov/user-api/internal/http/handler/middleware"
	"github.com/dgdraganov/user-api/internal/http/payload"
	"github.com/dgdraganov/user-api/internal/http/server"
	"github.com/dgdraganov/user-api/internal/minio"
	"github.com/dgdraganov/user-api/internal/rabbit"
	"github.com/dgdraganov/user-api/internal/repository"
	"github.com/dgdraganov/user-api/pkg/jwt"
	"github.com/dgdraganov/user-api/pkg/log"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger := log.NewZapLogger("user-api", zapcore.InfoLevel)

	config, err := config.NewAppConfig()
	if err != nil {
		logger.Errorw("failed to create config", "error", err)
		os.Exit(1)
	}

	// mySQL connection
	dbConn, err := db.NewMySqlDB(config.DBConnectionString)
	if err != nil {
		logger.Errorw("failed to connect to database", "error", err)
		os.Exit(1)
	}

	//rabbitMQ connection
	rbMQ, err := rabbit.NewRabbit(config.RabbitConnectionString)
	if err != nil {
		logger.Errorw("failed to connect to rabbitMQ", "error", err)
		os.Exit(1)
	}

	// jwt service
	jwtService := jwt.NewJWTService([]byte(config.JWTSecret))

	// repository
	repo := repository.NewUserRepository(dbConn)

	err = repo.MigrateTables(
		&repository.FileMetadata{},
		&repository.User{},
	)
	if err != nil {
		logger.Errorw("failed to migrate tables to database", "error", err)
		os.Exit(1)
	}

	err = repo.SeedUserTable(context.Background())
	if err != nil {
		logger.Errorw("failed to seed user table", "error", err)
		os.Exit(1)
	}

	// minio client
	minioClient, err := minio.NewMinioClient(config.MinioEndpoint, config.MinioAccessKey, config.MinioSecretKey)
	if err != nil {
		logger.Errorw("failed to create minio client", "error", err)
		os.Exit(1)
	}

	// ensure bucket exists
	err = minioClient.CreateBucket(context.Background(), config.MinioBucketName)
	if err != nil {
		logger.Errorw("failed to create minio bucket", "error", err)
		os.Exit(1)
	}

	// core service
	userService := core.NewUserService(
		logger,
		repo,
		rbMQ,
		jwtService,
		minioClient,
		config.MinioBucketName,
	)

	usrHandler := handler.NewUserHandler(
		logger,
		payload.DecodeValidator{},
		userService,
	)

	// middleware
	mux := http.NewServeMux()
	hdlr := middleware.NewLoggingMiddleware(logger).Logging(mux)
	hdlr = middleware.NewRequestIDMiddleware().RequestID(hdlr)

	// register routes
	mux.HandleFunc(handler.Authenticate, usrHandler.HandleAuthenticate)
	mux.HandleFunc(handler.ListUsers, usrHandler.HandleListUsers)
	mux.HandleFunc(handler.UploadFile, usrHandler.HandleFileUpload)
	mux.HandleFunc(handler.GetUser, usrHandler.HandleGetUser)
	mux.HandleFunc(handler.UserRegister, usrHandler.HandleRegisterUser)
	mux.HandleFunc(handler.UserUpdate, usrHandler.HandleUpdateUser)
	mux.HandleFunc(handler.UserDelete, usrHandler.HandleDeleteUser)

	srv := server.NewHTTP(logger, hdlr, config.Port)
	if err := run(srv); err != nil {
		logger.Errorw("server exited with an error", "error", err)
		os.Exit(1)
	}
}

func run(server *server.HTTPServer) error {
	// expect a signal to gracefully shutdown the server
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	errChan := server.Run()

	var err error
	select {
	case <-sig:
	case err = <-errChan:
	}

	sdErr := server.Shutdown()
	if err == http.ErrServerClosed && sdErr != nil {
		return fmt.Errorf("server shutdown: %w", sdErr)
	}

	return err
}
