package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bllooop/pvzservice/internal/usecase"
	mock_usecase "github.com/bllooop/pvzservice/internal/usecase/mocks"
	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_authIdentity(t *testing.T) {
	type mockBehavior func(r *mock_usecase.MockAuthorization, token string)

	testTable := []struct {
		name                 string
		headerName           string
		headerValue          string
		token                string
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "Ok",
			headerName:  "Authorization",
			headerValue: "Bearer token",
			token:       "token",
			mockBehavior: func(r *mock_usecase.MockAuthorization, token string) {
				r.EXPECT().ParseToken(token).Return("1", 1, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: "1",
		},
		{
			name:                 "Некорректное значение заголовка",
			headerName:           "",
			headerValue:          "Bearer token",
			token:                "token",
			mockBehavior:         func(r *mock_usecase.MockAuthorization, token string) {},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"message":"Пустой заголовок авторизации"}`,
		},
		{
			name:                 "Пустой токен",
			headerName:           "Authorization",
			headerValue:          "Bearer ",
			token:                "token",
			mockBehavior:         func(r *mock_usecase.MockAuthorization, token string) {},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"message":"Токен пуст"}`,
		},
		{
			name:        "Ошибка выдачи токена",
			headerName:  "Authorization",
			headerValue: "Bearer token",
			token:       "token",
			mockBehavior: func(r *mock_usecase.MockAuthorization, token string) {
				r.EXPECT().ParseToken(token).Return("", 0, errors.New("Некорректный ввод токена"))
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"message":"Некорректный ввод токена"}`,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_usecase.NewMockAuthorization(c)
			test.mockBehavior(repo, test.token)

			usecases := &usecase.Usecase{Authorization: repo}
			handler := Handler{Usecases: usecases}

			r := gin.New()
			r.GET("/identity", handler.authIdentity, func(c *gin.Context) {
				id, _ := c.Get(userCtx)
				c.String(http.StatusOK, "%d", id)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/identity", nil)
			req.Header.Set(test.headerName, test.headerValue)

			r.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.JSONEq(t, test.expectedResponseBody, w.Body.String())
		})
	}
}

func TestGetUserRole(t *testing.T) {
	testTable := []struct {
		name       string
		ctx        *gin.Context
		role       int
		shouldFail bool
	}{
		{
			name: "Ok",
			ctx: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Set(userCtx, 1)
				return c
			}(),
			role: 1,
		},
		{
			name: "Пусто",
			ctx: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				return c
			}(),
			shouldFail: true,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			id, err := getUserRole(test.ctx)
			if test.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.role, id)
		})
	}
}
