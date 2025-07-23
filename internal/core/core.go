package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	tokenIssuer "github.com/dgdraganov/user-api/pkg/jwt"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"

	"github.com/dgdraganov/user-api/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	rabbitExchange = "users"
)

// UserService is a struct that provides methods to interact with the Ethereum node and the database.
type UserService struct {
	logs       *zap.SugaredLogger
	repo       Repository
	rabbit     MessageBroker
	jwtIssuer  JWTIssuer
	minio      BlobStorage
	bucketName string
}

// NewUserService is a constructor function for the UserService type.
func NewUserService(logger *zap.SugaredLogger, repo Repository, rabbit MessageBroker, jwt JWTIssuer, minio BlobStorage, bucketName string) *UserService {
	return &UserService{
		logs:       logger,
		repo:       repo,
		rabbit:     rabbit,
		jwtIssuer:  jwt,
		minio:      minio,
		bucketName: bucketName,
	}
}

// Authenticate checks the provided email and password against the database. If the credentials are valid, it generates a JWT token for the user.
func (f *UserService) Authenticate(ctx context.Context, msg AuthMessage) (string, error) {
	user := repository.User{}
	err := f.repo.GetUserByEmail(ctx, msg.Email, &user)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrUserNotFound
		}
		return "", fmt.Errorf("get user from db: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(msg.Password)); err != nil {
		return "", ErrIncorrectPassword
	}

	tokenInfo := tokenIssuer.TokenInfo{
		Email:      user.Email,
		Subject:    user.ID,
		Role:       string(user.Role),
		Expiration: 24,
	}
	token := f.jwtIssuer.Generate(tokenInfo)
	signed, err := f.jwtIssuer.Sign(token)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return signed, nil
}

func (f *UserService) ValidateToken(ctx context.Context, token string) (jwt.MapClaims, error) {
	claims, err := f.jwtIssuer.Validate(token)
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}
	return claims, nil
}

func (f *UserService) ListUsers(ctx context.Context, page int, pageSize int) ([]UserRecord, error) {
	users := []repository.User{}
	err := f.repo.ListUsersByPage(ctx, page, pageSize, &users)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	userRecords := toUserRecordList(users)
	return userRecords, nil
}

func (f *UserService) UploadUserFile(ctx context.Context, objectName string, file io.Reader, fileSize int64) error {
	err := f.minio.UploadFile(ctx, f.bucketName, objectName, file, fileSize)
	if err != nil {
		return fmt.Errorf("upload file to bucket: %w", err)
	}
	return nil
}

func (f *UserService) SaveFileMetadata(ctx context.Context, fileName, bucket, userID string) error {
	fileMetadata := repository.FileMetadata{
		ID:         uuid.NewString(),
		FileName:   fileName,
		BucketName: bucket,
		UserID:     userID,
	}
	err := f.repo.SaveFileMetadata(ctx, fileMetadata)
	if err != nil {
		return fmt.Errorf("save file metadata: %w", err)
	}
	return nil
}

func (f *UserService) GetUser(ctx context.Context, id string) (UserRecord, error) {
	user := repository.User{}
	err := f.repo.GetUserByID(ctx, id, &user)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return UserRecord{}, ErrUserNotFound
		}
		return UserRecord{}, fmt.Errorf("get user from db: %w", err)
	}

	return toUserRecord(user), nil
}

func (f *UserService) RegisterUser(ctx context.Context, msg RegisterMessage) error {
	existingUser := repository.User{}
	err := f.repo.GetUserByEmail(ctx, msg.Email, &existingUser)
	if err != nil {
		if !errors.Is(err, repository.ErrUserNotFound) {
			return fmt.Errorf("check if user exists: %w", err)
		}
	}

	if existingUser.ID != "" {
		return ErrUserAlreadyExists
	}

	passHash, err := hashPassword(msg.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user := repository.User{
		ID:           uuid.NewString(),
		FirstName:    msg.FirstName,
		LastName:     msg.LastName,
		Email:        msg.Email,
		Age:          msg.Age,
		Role:         repository.RoleUser, // this is default role
		PasswordHash: passHash,
	}

	err = f.repo.CreateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("create user in db: %w", err)
	}
	return nil
}

func (f *UserService) UpdateUser(ctx context.Context, msg UpdateUserMessage, userID string) error {
	var user repository.User
	err := f.repo.GetUserByID(ctx, userID, &user)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("get user by id: %w", err)
	}

	userUpdate := msg.ToUser(user)
	err = f.repo.UpdateUser(ctx, userUpdate)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (f *UserService) DeleteUser(ctx context.Context, userID string) error {
	var user repository.User
	err := f.repo.GetUserByID(ctx, userID, &user)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("get user by id: %w", err)
	}

	err = f.repo.DeleteUser(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (f *UserService) PublishEvent(ctx context.Context, routingKey string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}

	err = f.rabbit.Publish(ctx, rabbitExchange, routingKey, body)
	if err != nil {
		return fmt.Errorf("publish event: %w", err)
	}
	return nil
}

func (f *UserService) ListUserFiles(ctx context.Context, userGUID string) ([]FileRecord, error) {
	var files []repository.FileMetadata
	err := f.repo.ListFilesByUserID(ctx, userGUID, &files)
	if err != nil {
		return nil, fmt.Errorf("list user files: %w", err)
	}

	fileRecs := toFileRecordList(files)

	return fileRecs, nil
}

func (f *UserService) DeleteUserFiles(ctx context.Context, userGUID string) error {
	var files []repository.FileMetadata
	err := f.repo.ListFilesByUserID(ctx, userGUID, &files)
	if err != nil {
		return fmt.Errorf("list user files: %w", err)
	}

	for _, file := range files {
		fileName := fmt.Sprintf("%s/%s", userGUID, file.FileName)
		err = f.minio.DeleteFile(ctx, f.bucketName, fileName)
		if err != nil {
			return fmt.Errorf("delete file from bucket: %w", err)
		}
	}

	err = f.repo.DeleteUserFiles(ctx, userGUID)
	if err != nil {
		return fmt.Errorf("delete user files: %w", err)
	}
	return nil
}

func toUserRecord(u repository.User) UserRecord {
	return UserRecord{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Role:      string(u.Role),
		Age:       u.Age,
	}
}

func toUserRecordList(users []repository.User) []UserRecord {
	userRecords := make([]UserRecord, 0, len(users))
	for _, u := range users {
		userRecords = append(userRecords, toUserRecord(u))
	}
	return userRecords
}

func toFileRecord(f repository.FileMetadata) FileRecord {
	return FileRecord{
		ID:       f.ID,
		FileName: f.FileName,
		Bucket:   f.BucketName,
	}
}

func toFileRecordList(files []repository.FileMetadata) []FileRecord {
	fileRecords := make([]FileRecord, 0, len(files))
	for _, file := range files {
		fileRecords = append(fileRecords, toFileRecord(file))
	}
	return fileRecords
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
