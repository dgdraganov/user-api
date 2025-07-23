package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/dgdraganov/user-api/internal/core"
	"github.com/dgdraganov/user-api/internal/http/handler"
	"github.com/dgdraganov/user-api/internal/http/handler/fake"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("UserHandler", func() {
	var (
		uHandler      *handler.UserHandler
		fakeService   *fake.CoreService
		fakeValidator *fake.RequestValidator
		logger        *zap.SugaredLogger
		recorder      *httptest.ResponseRecorder
		req           *http.Request
		testToken     string
		fakeErr       error
	)

	BeforeEach(func() {
		testToken = "test-token"
		fakeErr = errors.New("fake error")
		logger = zap.NewNop().Sugar()
		fakeService = new(fake.CoreService)
		fakeValidator = new(fake.RequestValidator)
		recorder = httptest.NewRecorder()
		uHandler = handler.NewUserHandler(logger, fakeValidator, fakeService)
	})

	Describe("HandleAuthenticate", func() {
		BeforeEach(func() {
			body := strings.NewReader(`{"email":"test@example.com","password":"pass123"}`)
			req = httptest.NewRequest(http.MethodPost, "/api/auth", body)
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
				fakeService.AuthenticateReturns("", core.ErrIncorrectPassword)
			})

			It("should return 401 Unauthorized", func() {
				uHandler.HandleAuthenticate(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("HandleRegisterUser", func() {
		BeforeEach(func() {
			body := bytes.NewBufferString(`{"email":"john@example.com","password":"pass123","first_name":"John","last_name":"Doe","age":25}`)
			req = httptest.NewRequest(http.MethodPost, "/api/users", body)
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
				fakeService.RegisterUserReturns(core.ErrUserAlreadyExists)
			})

			It("should return 409 Conflict", func() {
				Expect(fakeService.RegisterUserCallCount()).To(Equal(1))
				Expect(fakeService.PublishEventCallCount()).To(Equal(0))
				Expect(recorder.Code).To(Equal(http.StatusConflict))
			})
		})
	})

	Describe("HandleUpdateUser", func() {
		var (
			// updateReq payload.UpdateUserRequest
			userID = uuid.New().String()
		)

		BeforeEach(func() {
			updateReq := map[string]any{
				"first_name": "Test",
				"last_name":  "Test",
				"email":      "updated@example.com",
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

		It("updates user successfully", func() {
			Expect(recorder.Code).To(Equal(http.StatusOK))
			Expect(fakeService.UpdateUserCallCount()).To(Equal(1))
			_, updateMessage, userID := fakeService.UpdateUserArgsForCall(0)
			Expect(userID).To(Equal(userID))
			Expect(updateMessage.FirstName).To(Equal("Test"))
			Expect(updateMessage.LastName).To(Equal("Test"))
			Expect(updateMessage.Email).To(Equal("updated@example.com"))
			Expect(updateMessage.Age).To(Equal(30))
			Expect(fakeService.PublishEventCallCount()).To(Equal(1))
		})

		When("user is not authorized", func() {
			BeforeEach(func() {
				fakeService.ValidateTokenReturns(jwt.MapClaims{"sub": "another-user-id", "role": "user"}, nil)
			})

			It("should return 403 Forbidden", func() {
				Expect(recorder.Code).To(Equal(http.StatusForbidden))
				Expect(fakeService.UpdateUserCallCount()).To(Equal(0))
				Expect(fakeService.PublishEventCallCount()).To(Equal(0))
			})
		})

		When("validation fails", func() {
			BeforeEach(func() {
				fakeValidator.DecodeAndValidateJSONPayloadReturns(fakeErr)
			})

			It("should return 400 Bad Request", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(fakeService.UpdateUserCallCount()).To(Equal(0))
				Expect(fakeService.PublishEventCallCount()).To(Equal(0))
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
			})
		})
	})
})
