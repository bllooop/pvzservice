package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bllooop/pvzservice/internal/domain"
	"github.com/bllooop/pvzservice/internal/usecase"
	mock_usecase "github.com/bllooop/pvzservice/internal/usecase/mocks"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_dummyLogin(t *testing.T) {
	type mockBehavior func(s *mock_usecase.MockAuthorization, role string)
	testTable := []struct {
		name                 string
		inputBody            string
		role                 string
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "OK",
			inputBody: `{"role":"moderator"}`,
			role:      "moderator",
			mockBehavior: func(s *mock_usecase.MockAuthorization, role string) {
				s.EXPECT().GenerateToken(uuid.MustParse("22222222-2222-2222-2222-222222222222"), 2).Return("valid.jwt.token", nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"token":"valid.jwt.token"}`,
		},
		{
			name:                 "Invalid JSON Input",
			inputBody:            `{"role":1000}`,
			mockBehavior:         func(s *mock_usecase.MockAuthorization, role string) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Неверный запрос"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_usecase.NewMockAuthorization(c)
			testCase.mockBehavior(repo, testCase.role)

			usecases := &usecase.Usecase{Authorization: repo}
			handler := Handler{Usecases: usecases}
			r := gin.New()
			r.POST("/dummyLogin", handler.DummyLogin)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)

			if json.Valid([]byte(testCase.expectedResponseBody)) {
				assert.JSONEq(t, testCase.expectedResponseBody, w.Body.String())
			} else {
				assert.Equal(t, testCase.expectedResponseBody, strings.TrimSpace(w.Body.String()))
			}
		})
	}
}

func TestHandler_signUp(t *testing.T) {
	type mockBehavior func(s *mock_usecase.MockAuthorization, user domain.User)

	testTable := []struct {
		name                 string
		inputBody            string
		inputUser            domain.User
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "OK",
			inputBody: `{"email":"test", "password":"12345", "role":"employee"}`,
			inputUser: domain.User{
				Email:    "test",
				Password: "12345",
				Role:     "employee",
			},
			mockBehavior: func(s *mock_usecase.MockAuthorization, user domain.User) {
				s.EXPECT().CreateUser(user).Return(1, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"id":1}`,
		},
		{
			name:      "Ошибка выполнения запроса",
			inputBody: `{"email": "test", "password":"12345","role":"moderator"}`,
			inputUser: domain.User{
				Email:    "test",
				Password: "12345",
				Role:     "moderator",
			},
			mockBehavior: func(s *mock_usecase.MockAuthorization, user domain.User) {
				s.EXPECT().CreateUser(user).Return(0, errors.New("Internal Server Error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Internal Server Error"}`,
		},
		{
			name:                 "Плохой ввод",
			inputBody:            `{"email":1000}`,
			inputUser:            domain.User{},
			mockBehavior:         func(s *mock_usecase.MockAuthorization, user domain.User) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Неверный запрос"}`,
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_usecase.NewMockAuthorization(c)
			testCase.mockBehavior(repo, testCase.inputUser)

			usecases := &usecase.Usecase{Authorization: repo}
			handler := Handler{Usecases: usecases}
			r := gin.New()
			api := r.Group("/api")
			api.POST("/register", handler.SignUp)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/register",
				bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)
			assert.Equal(t, w.Code, testCase.expectedStatusCode)
			if json.Valid([]byte(testCase.expectedResponseBody)) {
				assert.JSONEq(t, testCase.expectedResponseBody, w.Body.String())
			} else {
				assert.Equal(t, testCase.expectedResponseBody, strings.TrimSpace(w.Body.String()))
			}
		})
	}
}

func TestHandler_signIn(t *testing.T) {
	type mockBehavior func(s *mock_usecase.MockAuthorization, email, password string)
	userID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	testTable := []struct {
		name                 string
		inputBody            string
		email                string
		password             string
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "OK",
			inputBody: `{"email":"name", "password":"12345"}`,
			email:     "name",
			password:  "12345",
			mockBehavior: func(s *mock_usecase.MockAuthorization, email, password string) {
				s.EXPECT().SignUser("name", "12345").Return(domain.User{Id: userID, Role: "moderator"}, nil)
				s.EXPECT().GenerateToken(userID, 2).Return("valid.jwt.token", nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"token":"valid.jwt.token"}`,
		},
		{
			name:      "Пользователь не найден",
			inputBody: `{"email":"notname", "password":"password123"}`,
			email:     "name",
			password:  "password123",
			mockBehavior: func(s *mock_usecase.MockAuthorization, email, password string) {
				s.EXPECT().SignUser("notname", "password123").Return(domain.User{}, errors.New("пользователь не найден"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Ошибка авторизации: пользователь не найден"}`,
		},
		{
			name:                 "Invalid JSON Input",
			inputBody:            `{"email":1000}`,
			mockBehavior:         func(s *mock_usecase.MockAuthorization, email, password string) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Неверный запрос"}`,
		},
		{
			name:      "Ошибка авторизации",
			inputBody: `{"email":"test", "password":"12345","role":"employee"}`,
			email:     "test",
			password:  "12345",
			mockBehavior: func(s *mock_usecase.MockAuthorization, email, password string) {
				s.EXPECT().SignUser("test", "12345").Return(domain.User{}, errors.New("Internal Server Error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Ошибка авторизации: Internal Server Error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_usecase.NewMockAuthorization(c)
			testCase.mockBehavior(repo, testCase.email, testCase.password)

			usecases := &usecase.Usecase{Authorization: repo}
			handler := Handler{Usecases: usecases}
			r := gin.New()
			r.POST("/api/auth/sign-in", handler.SignIn)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/auth/sign-in", bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)

			if json.Valid([]byte(testCase.expectedResponseBody)) {
				assert.JSONEq(t, testCase.expectedResponseBody, w.Body.String())
			} else {
				assert.Equal(t, testCase.expectedResponseBody, strings.TrimSpace(w.Body.String()))
			}
		})
	}
}
