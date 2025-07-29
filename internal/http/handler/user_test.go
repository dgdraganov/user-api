package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/dgdraganov/user-api/internal/http/handler"
	"github.com/dgdraganov/user-api/internal/http/handler/fake"
	"github.com/dgdraganov/user-api/internal/http/payload"
	"github.com/dgdraganov/user-api/internal/service"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("UserHandler", func() {
	var (
		uHandler      *handler.UserHandler
		fakeService   *fake.UserService
		fakeValidator *fake.RequestValidator
		logger        *zap.SugaredLogger
		recorder      *httptest.ResponseRecorder
		req           *http.Request
		testToken     string
		fakeErr       error
		userID        string
	)

	BeforeEach(func() {
		testToken = "test-token"
		userID = uuid.New().String()
		fakeErr = errors.New("fake error")
		logger = zap.NewNop().Sugar()
		fakeService = new(fake.UserService)
		fakeValidator = new(fake.RequestValidator)
		recorder = httptest.NewRecorder()
		uHandler = handler.NewUserHandler(logger, fakeValidator, fakeService)
	})

	Describe("HandleAuthenticate", func() {
		BeforeEach(func() {
			authReq := map[string]any{
				"password": "pass123",
				"email":    "test@example.com",
			}
			jsonBody, _ := json.Marshal(authReq)
			req = httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			fakeValidator.DecodeAndValidateJSONPayloadReturns(nil)
			fakeService.AuthenticateReturns(testToken, nil)
		})

		JustBeforeEach(func() {
			uHandler.HandleAuthenticate(recorder, req)
		})

		It("should return a token when credentials are valid", func() {
			Expect(recorder.Code).To(Equal(http.StatusOK))
			var resp map[string]string
			json.NewDecoder(recorder.Body).Decode(&resp)
			Expect(resp["token"]).To(Equal(testToken))
		})

		When("decode or validate fails", func() {
			BeforeEach(func() {
				fakeValidator.DecodeAndValidateJSONPayloadReturns(fakeErr)
			})

			It("should return 400 bad request", func() {
				uHandler.HandleAuthenticate(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		When("credentials are invalid", func() {
			BeforeEach(func() {
				fakeService.AuthenticateReturns("", service.ErrIncorrectPassword)
			})

			It("should return 401 Unauthorized", func() {
				uHandler.HandleAuthenticate(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("HandleRegisterUser", func() {
		BeforeEach(func() {
			regReq := map[string]any{
				"email":      "john@example.com",
				"password":   "pass123",
				"first_name": "John",
				"last_name":  "Doe",
				"age":        25,
			}
			jsonBody, _ := json.Marshal(regReq)
			req = httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			fakeValidator.DecodeAndValidateJSONPayloadStub = func(r *http.Request, obj any) error {
				return json.NewDecoder(r.Body).Decode(obj)
			}
			fakeService.RegisterUserReturns(nil)
			fakeService.PublishEventReturns(nil)
		})

		JustBeforeEach(func() {
			uHandler.HandleRegisterUser(recorder, req)
		})

		It("should return 201 Created when registration is successful", func() {
			Expect(recorder.Code).To(Equal(http.StatusCreated))
			Expect(fakeService.RegisterUserCallCount()).To(Equal(1))
			_, regMessage := fakeService.RegisterUserArgsForCall(0)
			Expect(regMessage.Email).To(Equal("john@example.com"))
			Expect(regMessage.Password).To(Equal("pass123"))
			Expect(regMessage.FirstName).To(Equal("John"))
			Expect(regMessage.LastName).To(Equal("Doe"))
			Expect(regMessage.Age).To(Equal(25))

			Expect(fakeService.PublishEventCallCount()).To(Equal(1))
			_, eventType, eventData := fakeService.PublishEventArgsForCall(0)
			Expect(eventType).To(Equal("user.event.registered"))
			Expect(eventData).To(HaveKeyWithValue("email", "john@example.com"))
			Expect(eventData).To(HaveKeyWithValue("first_name", "John"))
			Expect(eventData).To(HaveKeyWithValue("last_name", "Doe"))
			Expect(eventData).To(HaveKeyWithValue("age", 25))
		})

		When("user already exists", func() {
			BeforeEach(func() {
				fakeService.RegisterUserReturns(service.ErrUserAlreadyExists)
			})

			It("should return 409 Conflict", func() {
				Expect(fakeService.RegisterUserCallCount()).To(Equal(1))
				Expect(fakeService.PublishEventCallCount()).To(Equal(0))
				Expect(recorder.Code).To(Equal(http.StatusConflict))
			})
		})
	})

	Describe("HandleUpdateUser", func() {
		BeforeEach(func() {
			updateReq := map[string]any{
				"first_name": "Test",
				"last_name":  "Test",
				"email":      "test@example.com",
				"age":        30,
			}

			jsonBody, _ := json.Marshal(updateReq)
			req = httptest.NewRequest(http.MethodPut, "/api/users/"+userID, bytes.NewBuffer(jsonBody))
			req = req.WithContext(context.Background())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("AUTH_TOKEN", testToken)

			fakeValidator.DecodeAndValidateJSONPayloadStub = func(r *http.Request, obj any) error {
				return json.NewDecoder(r.Body).Decode(obj)
			}
			fakeService.ValidateTokenReturns(jwt.MapClaims{"sub": userID, "role": "user"}, nil)
			fakeService.UpdateUserReturns(nil)
			fakeService.PublishEventReturns(nil)
		})

		JustBeforeEach(func() {
			uHandler.HandleUpdateUser(recorder, req)
		})
		When("user is not admin", func() {
			BeforeEach(func() {
				fakeService.ValidateTokenReturns(jwt.MapClaims{"sub": "random-user-id", "role": "user"}, nil)
			})

			It("fails to update other user's profile", func() {
				Expect(recorder.Code).To(Equal(http.StatusForbidden))
				Expect(fakeService.UpdateUserCallCount()).To(Equal(0))
				Expect(fakeService.PublishEventCallCount()).To(Equal(0))
			})
		})

		When("user is admin", func() {
			var adminUserID string

			BeforeEach(func() {
				adminUserID = uuid.New().String()
				fakeService.ValidateTokenReturns(jwt.MapClaims{"sub": adminUserID, "role": "admin"}, nil)
			})

			It("updates other user's profile successfully", func() {
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(fakeService.UpdateUserCallCount()).To(Equal(1))
				_, updateMessage, userGUID := fakeService.UpdateUserArgsForCall(0)
				Expect(userGUID).To(Equal(userID))
				Expect(userID).ToNot(Equal(adminUserID))
				Expect(updateMessage.FirstName).To(Equal("Test"))
				Expect(updateMessage.LastName).To(Equal("Test"))
				Expect(updateMessage.Email).To(Equal("test@example.com"))
				Expect(updateMessage.Age).To(Equal(30))
				Expect(fakeService.PublishEventCallCount()).To(Equal(1))
			})
		})

		When("payload has invalid fields", func() {
			BeforeEach(func() {
				fakeValidator.DecodeAndValidateJSONPayloadReturns(fakeErr)
			})

			It("should return 400 Bad Request", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(fakeService.UpdateUserCallCount()).To(Equal(0))
				Expect(fakeService.PublishEventCallCount()).To(Equal(0))
				Expect(fakeService.ValidateTokenCallCount()).To(Equal(1))
			})
		})

		When("update fails internally", func() {
			BeforeEach(func() {
				fakeService.UpdateUserReturns(errors.New("db error"))
			})

			It("should return 500 Internal Server Error", func() {
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(fakeService.UpdateUserCallCount()).To(Equal(1))
				Expect(fakeService.PublishEventCallCount()).To(Equal(0))
				Expect(fakeService.ValidateTokenCallCount()).To(Equal(1))
			})
		})

		When("token validation fails", func() {
			BeforeEach(func() {
				fakeService.ValidateTokenReturns(nil, fakeErr)
			})

			It("should return 401 Unauthorized", func() {
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				Expect(fakeService.UpdateUserCallCount()).To(Equal(0))
				Expect(fakeService.PublishEventCallCount()).To(Equal(0))
				Expect(fakeService.ValidateTokenCallCount()).To(Equal(1))
				_, token := fakeService.ValidateTokenArgsForCall(0)
				Expect(token).To(Equal(testToken))
			})
		})
	})

	Describe("HandleListUsers", func() {
		var (
			queryParams = url.Values{}
			pageVal     int
			pageSizeVal int
		)

		BeforeEach(func() {
			pageVal = 1
			pageSizeVal = 10
			queryParams.Set("page", fmt.Sprintf("%d", pageVal))
			queryParams.Set("page_size", fmt.Sprintf("%d", pageSizeVal))
			req = httptest.NewRequest(http.MethodGet, "/api/users?"+queryParams.Encode(), nil)
			req.Header.Set("AUTH_TOKEN", testToken)

			fakeValidator.DecodeAndValidateQueryParamsStub = func(r *http.Request, u payload.URLDecoder) error {
				return u.DecodeFromURLValues(r.URL.Query())
			}

			fakeService.ValidateTokenReturns(jwt.MapClaims{"sub": userID, "role": "user"}, nil)
			fakeService.ListUsersReturns([]service.UserRecord{
				{
					ID:        uuid.New().String(),
					Email:     "alice@example.com",
					FirstName: "Alice",
					LastName:  "Cooper",
					Age:       30,
				},
				{
					ID:        uuid.New().String(),
					Email:     "bob@example.com",
					FirstName: "Bob",
					LastName:  "Marley",
					Age:       25,
				},
			}, nil)
		})

		JustBeforeEach(func() {
			uHandler.HandleListUsers(recorder, req)
		})

		It("should return a list of users", func() {
			Expect(recorder.Code).To(Equal(http.StatusOK))
			Expect(fakeService.ListUsersCallCount()).To(Equal(1))
			_, page, pageSize := fakeService.ListUsersArgsForCall(0)
			Expect(page).To(Equal(pageVal))
			Expect(pageSize).To(Equal(pageSizeVal))
		})

		When("token validation fails", func() {
			BeforeEach(func() {
				fakeService.ValidateTokenReturns(nil, fakeErr)
			})

			It("should return 401 Unauthorized", func() {
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
				Expect(fakeService.ListUsersCallCount()).To(Equal(0))
			})
		})

		When("listing users fails", func() {
			BeforeEach(func() {
				fakeService.ListUsersReturns(nil, fakeErr)
			})

			It("should return 500 Internal Server Error", func() {
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(fakeService.ListUsersCallCount()).To(Equal(1))
			})
		})
	})

	Describe("HandleGetUser", func() {

		var regularUserID string

		BeforeEach(func() {
			regularUserID = uuid.New().String()
			req = httptest.NewRequest(http.MethodGet, "/api/users/"+userID, nil)
			req.Header.Set("AUTH_TOKEN", testToken)

			fakeService.ValidateTokenReturns(jwt.MapClaims{"sub": regularUserID, "role": "user"}, nil)
			fakeService.GetUserReturns(service.UserRecord{
				ID:        userID,
				Email:     "john@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Age:       25,
				Role:      "admin",
			}, nil)
		})
		JustBeforeEach(func() {
			uHandler.HandleGetUser(recorder, req)
		})
		It("should return user details", func() {
			Expect(recorder.Code).To(Equal(http.StatusOK))
			var userData map[string]any
			err := json.NewDecoder(recorder.Body).Decode(&userData)
			Expect(err).ToNot(HaveOccurred())
			user, ok := userData["user"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(user["id"]).To(Equal(userID))
			Expect(user["email"]).To(Equal("john@example.com"))
			Expect(user["first_name"]).To(Equal("John"))
			Expect(user["last_name"]).To(Equal("Doe"))
			age := fmt.Sprintf("%v", user["age"])
			Expect(age).To(Equal("25"))
		})
	})
})
