package service_test

import (
	"context"
	"errors"

	"github.com/dgdraganov/user-api/internal/repository"
	"github.com/dgdraganov/user-api/internal/service"
	"github.com/dgdraganov/user-api/internal/service/fake"
	tokenIssuer "github.com/dgdraganov/user-api/pkg/jwt"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"go.uber.org/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserService", func() {
	var (
		fakeRepo   *fake.Repository
		fakeJWT    *fake.JWTIssuer
		fakeMinio  *fake.BlobStorage
		fakeRabbit *fake.MessageBroker
		fakeLogger *zap.SugaredLogger
		ctx        context.Context

		usrSvc *service.UserService

		fakeErr error
	)

	BeforeEach(func() {
		fakeRepo = new(fake.Repository)
		fakeJWT = new(fake.JWTIssuer)
		fakeMinio = new(fake.BlobStorage)
		fakeRabbit = new(fake.MessageBroker)
		bucketName := "test-bucket"
		fakeLogger = zap.NewNop().Sugar()
		ctx = context.Background()
		fakeErr = errors.New("fake error")

		usrSvc = service.NewUserService(fakeLogger, fakeRepo, fakeRabbit, fakeJWT, fakeMinio, bucketName)
	})

	Describe("Authenticate", func() {
		var (
			authMsg        service.AuthMessage
			token          string
			err            error
			userId         string
			tokenInfo      tokenIssuer.TokenInfo
			hashedPassword string
			genToken       *jwt.Token
		)

		BeforeEach(func() {
			userId = uuid.New().String()
			// hashed password for "testpass" using bcrypt
			hashedPassword = "$2a$10$1MZHKX./8Dxi9t.F1/gnx.njCcEty299Hx01GLEms2moa3brpT0ky"
			genToken = jwt.New(jwt.SigningMethodHS256)

			authMsg = service.AuthMessage{
				Email:    "testuser",
				Password: "testpass",
			}

			tokenInfo = tokenIssuer.TokenInfo{
				Email:      authMsg.Email,
				Subject:    userId,
				Expiration: 24,
			}
		})

		JustBeforeEach(func() {
			token, err = usrSvc.Authenticate(ctx, authMsg)
		})

		When("user exists and password matches", func() {
			BeforeEach(func() {
				fakeRepo.GetUserByEmailStub = func(ctx context.Context, email string, user *repository.User) error {
					*user = repository.User{
						Email:        authMsg.Email,
						PasswordHash: hashedPassword,
						ID:           userId,
					}
					return nil
				}

				fakeJWT.GenerateReturns(genToken)
				fakeJWT.SignReturns("signed.token", nil)

			})

			It("should return a signed token", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(token).To(Equal("signed.token"))

				Expect(fakeRepo.GetUserByEmailCallCount()).To(Equal(1))
				_, email, _ := fakeRepo.GetUserByEmailArgsForCall(0)
				Expect(email).To(Equal(authMsg.Email))

				Expect(fakeJWT.GenerateCallCount()).To(Equal(1))
				argGen := fakeJWT.GenerateArgsForCall(0)
				Expect(argGen).To(Equal(tokenInfo))

				Expect(fakeJWT.SignCallCount()).To(Equal(1))
				argSign := fakeJWT.SignArgsForCall(0)
				Expect(argSign).To(Equal(genToken))
			})
		})

		When("user does not exist", func() {
			BeforeEach(func() {
				fakeRepo.GetUserByEmailStub = func(ctx context.Context, email string, user *repository.User) error {
					return repository.ErrUserNotFound
				}
			})

			It("should return user not found error", func() {
				Expect(err).To(MatchError(service.ErrUserNotFound))
			})
		})

		When("password does not match", func() {
			BeforeEach(func() {
				fakeRepo.GetUserByEmailStub = func(ctx context.Context, email string, user *repository.User) error {
					*user = repository.User{
						Email:        authMsg.Email,
						PasswordHash: hashedPassword,
					}
					return nil
				}
				authMsg.Password = "wrongpass"
			})

			It("should return incorrect password error", func() {
				Expect(err).To(MatchError(service.ErrIncorrectPassword))
			})
		})

		When("token signing fails", func() {
			BeforeEach(func() {
				fakeRepo.GetUserByEmailStub = func(ctx context.Context, email string, user *repository.User) error {
					*user = repository.User{
						Email:        authMsg.Email,
						PasswordHash: hashedPassword,
						ID:           userId,
					}
					return nil
				}
				fakeJWT.SignReturns("", fakeErr)
			})

			It("should return signing error", func() {
				Expect(err).To(MatchError(fakeErr))
			})
		})
	})

})
