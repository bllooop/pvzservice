package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bllooop/pvzservice/internal/domain"
	"github.com/bllooop/pvzservice/internal/usecase"
	mock_usecase "github.com/bllooop/pvzservice/internal/usecase/mocks"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_createPvz(t *testing.T) {
	type mockBehavior func(s *mock_usecase.MockPvz, pvz domain.PVZ)
	fixedTime := time.Date(2025, 4, 10, 15, 5, 17, 329922000, time.UTC)
	userID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	testTable := []struct {
		name                 string
		inputBody            string
		inputPVZ             domain.PVZ
		inputUserRole        int
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "OK",
			inputBody: `{"city":"Москва"}`,
			inputPVZ: domain.PVZ{
				DateRegister: &fixedTime,
				City:         "Москва",
			},
			inputUserRole: 2,
			mockBehavior: func(s *mock_usecase.MockPvz, pvz domain.PVZ) {
				s.EXPECT().CreatePvz(pvz).Return(domain.PVZ{
					Id:           &userID,
					DateRegister: &fixedTime,
					City:         "Москва",
				}, nil)
			},
			expectedStatusCode: 200,
			expectedResponseBody: fmt.Sprintf(`{
				"data": {
					"city": "Москва", 
					"id": "%s", 
					"registrationDate": "2025-04-10T15:05:17.329922Z"
				},
				"message": "ПВЗ создан"
			}`, userID.String()),
		},
		{
			name:      "Ошибка выполнения запроса",
			inputBody: `{"city":"Москва"}`,
			inputPVZ: domain.PVZ{
				DateRegister: &fixedTime,
				City:         "Москва",
			},
			inputUserRole: 2,
			mockBehavior: func(s *mock_usecase.MockPvz, pvz domain.PVZ) {
				s.EXPECT().CreatePvz(pvz).Return(domain.PVZ{}, errors.New("Internal Server Error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Ошибка выполнения запроса Internal Server Error"}`,
		},
		{
			name:          "Запрещен доступ",
			inputBody:     `{"city":"Москва"}`,
			inputPVZ:      domain.PVZ{},
			inputUserRole: 1,
			mockBehavior: func(s *mock_usecase.MockPvz, pvz domain.PVZ) {
				s.EXPECT().CreatePvz(gomock.Any()).Times(0)
			},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Доступ запрещен"}`,
		},
		{
			name:      "Плохой ввод",
			inputBody: `{"city":10000}`,
			inputPVZ: domain.PVZ{
				DateRegister: &fixedTime,
				City:         "1000",
			},
			inputUserRole:        2,
			mockBehavior:         func(s *mock_usecase.MockPvz, pvz domain.PVZ) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Неверный запрос"}`,
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_usecase.NewMockPvz(c)
			testCase.mockBehavior(repo, testCase.inputPVZ)

			usecases := &usecase.Usecase{Pvz: repo}
			handler := Handler{
				Usecases: usecases,
				Now:      func() time.Time { return fixedTime },
			}
			r := gin.New()
			api := r.Group("/api")
			api.POST("/pvz", func(c *gin.Context) {
				c.Set("userRole", testCase.inputUserRole)
				handler.CreatePvz(c)
			})
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/pvz",
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

func TestHandler_closeLast(t *testing.T) {
	type mockBehavior func(s *mock_usecase.MockPvz, pvzId uuid.UUID)
	fixedTime := time.Date(2025, 4, 10, 15, 5, 17, 329922000, time.UTC)
	userID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	stat := "close"

	testTable := []struct {
		name                 string
		inputPvzId           string
		inputUserRole        int
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:          "OK",
			inputPvzId:    userID.String(),
			inputUserRole: 1,
			mockBehavior: func(s *mock_usecase.MockPvz, pvzId uuid.UUID) {
				s.EXPECT().CloseReception(pvzId).Return(domain.ProductReception{
					Id:           &userID,
					DateReceived: &fixedTime,
					PVZId:        &pvzId,
					Status:       &stat,
				}, nil)
			},
			expectedStatusCode: 200,
			expectedResponseBody: fmt.Sprintf(`{
                "data": {
                    "id": "%s",
                    "dateTime": "2025-04-10T15:05:17.329922Z",
                    "pvzId": "%s",
                    "status": "close"
                },
                "message": "Приемка закрыта"
            }`, userID.String(), userID.String()),
		},
		{
			name:          "Ошибка выполнения запроса",
			inputPvzId:    userID.String(),
			inputUserRole: 1,
			mockBehavior: func(s *mock_usecase.MockPvz, pvzId uuid.UUID) {
				s.EXPECT().CloseReception(pvzId).Return(domain.ProductReception{}, errors.New("Internal Server Error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Ошибка выполнения запроса Internal Server Error"}`,
		},
		/*	{
			name:          "Ошибка получения роли",
			inputPvzId:    userID.String(),
			inputUserRole: 1,
			mockBehavior: func(s *mock_usecase.MockPvz, pvzId uuid.UUID) {
				s.EXPECT().GetUserRole(pvzId).Return(0, errors.New("Ошибка базы данных"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Ошибка получения роли Ошибка базы данных"}`,
		},*/
		{
			name:                 "Запрещен доступ",
			inputPvzId:           userID.String(),
			inputUserRole:        2,
			mockBehavior:         nil,
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Доступ запрещен"}`,
		},
		{
			name:                 "Пустой параметр pvzId",
			inputPvzId:           "invalid-uuid",
			inputUserRole:        1,
			mockBehavior:         nil,
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Некорректный UUID ПВЗ"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_usecase.NewMockPvz(c)
			var parsedPvzId uuid.UUID
			var parseErr error
			if testCase.inputPvzId != "" && testCase.inputPvzId != "invalid-uuid" {
				parsedPvzId, parseErr = uuid.Parse(testCase.inputPvzId)
			}

			if parseErr == nil && testCase.mockBehavior != nil {
				testCase.mockBehavior(repo, parsedPvzId)
			}

			usecases := &usecase.Usecase{Pvz: repo}
			handler := Handler{
				Usecases: usecases,
				Now:      func() time.Time { return fixedTime },
			}
			r := gin.New()
			api := r.Group("/api")
			api.POST("/pvz/:pvzId/close_last_reception", func(c *gin.Context) {
				c.Set("userRole", testCase.inputUserRole)
				handler.CloseLast(c)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/pvz/"+testCase.inputPvzId+"/close_last_reception", nil)

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

func TestHandler_deleteLast(t *testing.T) {
	type mockBehavior func(s *mock_usecase.MockPvz, pvzId uuid.UUID)
	userID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	testTable := []struct {
		name                 string
		inputPvzId           string
		inputUserRole        int
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:          "OK",
			inputPvzId:    userID.String(),
			inputUserRole: 1,
			mockBehavior: func(s *mock_usecase.MockPvz, pvzId uuid.UUID) {
				s.EXPECT().DeleteLastProduct(pvzId).Return(nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{ "message": "Товар удален"}`,
		},
		{
			name:          "Ошибка выполнения запроса",
			inputPvzId:    userID.String(),
			inputUserRole: 1,
			mockBehavior: func(s *mock_usecase.MockPvz, pvzId uuid.UUID) {
				s.EXPECT().DeleteLastProduct(pvzId).Return(errors.New("Internal Server Error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Ошибка выполнения запроса Internal Server Error"}`,
		},
		{
			name:                 "Пустой параметр pvzId",
			inputPvzId:           "invalid-uuid",
			inputUserRole:        1,
			mockBehavior:         nil,
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Некорректный UUID ПВЗ"}`,
		},
		{
			name:                 "Запрещен доступ",
			inputPvzId:           userID.String(),
			inputUserRole:        2,
			mockBehavior:         nil,
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Доступ запрещен"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_usecase.NewMockPvz(c)
			var parsedPvzId uuid.UUID
			var parseErr error
			if testCase.inputPvzId != "" && testCase.inputPvzId != "invalid-uuid" {
				parsedPvzId, parseErr = uuid.Parse(testCase.inputPvzId)
			}

			if parseErr == nil && testCase.mockBehavior != nil {
				testCase.mockBehavior(repo, parsedPvzId)
			}

			usecases := &usecase.Usecase{Pvz: repo}
			handler := Handler{
				Usecases: usecases,
			}
			r := gin.New()
			api := r.Group("/api")
			api.POST("/pvz/:pvzId/delete_last_reception", func(c *gin.Context) {
				c.Set("userRole", testCase.inputUserRole)
				handler.DeleteLast(c)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/pvz/"+testCase.inputPvzId+"/delete_last_reception", nil)

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

func TestHandler_createReception(t *testing.T) {
	type mockBehavior func(s *mock_usecase.MockPvz, reception domain.ProductReception)
	fixedTime := time.Date(2025, 4, 10, 15, 5, 17, 329922000, time.UTC)
	userID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	stat := "in_progress"

	testTable := []struct {
		name                 string
		inputUserRole        int
		inputBody            string
		inputRecep           domain.ProductReception
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:          "OK",
			inputUserRole: 1,
			inputBody: fmt.Sprintf(`{
				"pvzId": "%s"
			}`, userID.String()),
			inputRecep: domain.ProductReception{
				DateReceived: &fixedTime,
				Status:       &stat,
				PVZId:        &userID,
			},
			mockBehavior: func(s *mock_usecase.MockPvz, reception domain.ProductReception) {
				reception.PVZId = &userID
				s.EXPECT().CreateRecep(reception).Return(domain.ProductReception{
					Id:           &userID,
					DateReceived: &fixedTime,
					PVZId:        &userID,
					Status:       &stat,
				}, nil)
			},
			expectedStatusCode: 200,
			expectedResponseBody: fmt.Sprintf(`{
				"data": {
					"id": "%s",  
					"dateTime": "2025-04-10T15:05:17.329922Z", 
					"pvzId": "%s", 
					"status": "in_progress"
				},
				"message": "Приемка создана"
			}`, userID.String(), userID.String()),
		},
		{
			name: "Ошибка выполнения запроса",
			inputBody: fmt.Sprintf(`{
				"pvzId": "%s"
			}`, userID.String()),
			inputRecep: domain.ProductReception{
				DateReceived: &fixedTime,
				Status:       &stat,
				PVZId:        &userID,
			},
			inputUserRole: 1,
			mockBehavior: func(s *mock_usecase.MockPvz, reception domain.ProductReception) {
				s.EXPECT().CreateRecep(reception).Return(domain.ProductReception{}, errors.New("Internal Server Error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Ошибка выполнения запроса Internal Server Error"}`,
		},
		{
			name: "Запрещен доступ",
			inputBody: fmt.Sprintf(`{
				"pvzId": "%s"
			}`, userID.String()),
			inputRecep: domain.ProductReception{
				DateReceived: &fixedTime,
				Status:       &stat,
				PVZId:        &userID,
			},
			inputUserRole: 2,
			mockBehavior: func(s *mock_usecase.MockPvz, pvz domain.ProductReception) {
				s.EXPECT().CreateRecep(gomock.Any()).Times(0)
			},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Доступ запрещен"}`,
		},
		/*	{
			name: "Пустой параметр pvzId",
			inputBody: `{
				"pvzId": "invalid-uuid",
			}`,
			inputRecep: domain.ProductReception{
				DateReceived: &fixedTime,
				Status:       &stat,
				PVZId:        nil,
			},
			inputUserRole:        1,
			mockBehavior:         nil,
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Некорректный UUID ПВЗ"}`,
		},*/
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_usecase.NewMockPvz(c)

			/*var parsedPvzId uuid.UUID
			var parseErr error
			if testCase.inputBody != "" && testCase.inputBody != "invalid-uuid" {
				parsedPvzId, parseErr = uuid.Parse(testCase.inputBody)
			}

			if parseErr == nil && testCase.mockBehavior != nil {
				reception := domain.ProductReception{
					PVZId:  &parsedPvzId,
					Status: &stat,
				}
				testCase.mockBehavior(repo, reception)
			}*/
			testCase.mockBehavior(repo, testCase.inputRecep)
			usecases := &usecase.Usecase{Pvz: repo}
			handler := Handler{
				Usecases: usecases,
				Now:      func() time.Time { return fixedTime },
			}
			r := gin.New()
			api := r.Group("/api")
			api.POST("/receptions", func(c *gin.Context) {
				c.Set("userRole", testCase.inputUserRole)
				handler.CreateReceptions(c)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/receptions", strings.NewReader(testCase.inputBody))
			req.Header.Set("Content-Type", "application/json")

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

func TestHandler_addProduct(t *testing.T) {
	type mockBehavior func(s *mock_usecase.MockPvz, product domain.Product)
	fixedTime := time.Date(2025, 4, 10, 15, 5, 17, 329922000, time.UTC)
	userID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	testTable := []struct {
		name                 string
		inputUserRole        int
		inputBody            string
		inputProd            domain.Product
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:          "OK",
			inputUserRole: 1,
			inputBody: fmt.Sprintf(`{
				"pvzId": "%s",
				"type":"электроника"
			}`, userID.String()),
			inputProd: domain.Product{
				DateReceived: &fixedTime,
				PVZId:        &userID,
				ReceptionId:  nil,
				Type:         "электроника",
			},
			mockBehavior: func(s *mock_usecase.MockPvz, product domain.Product) {
				s.EXPECT().AddProdToRecep(product).Return(domain.Product{
					Id:           &userID,
					DateReceived: &fixedTime,
					ReceptionId:  &userID,
					Type:         "электроника",
				}, nil)
			},
			expectedStatusCode: 200,
			expectedResponseBody: fmt.Sprintf(`{
				"data":     {
            "id": "%s",
            "dateTime": "2025-04-10T15:05:17.329922Z",
            "type": "электроника",
            "receptionId": "%s"
          },
				"message": "Товар добавлен"
			}`, userID.String(), userID.String()),
		},
		{
			name: "Ошибка выполнения запроса",
			inputBody: fmt.Sprintf(`{
				"pvzId": "%s",
				"type":"электроника"
			}`, userID.String()),
			inputProd: domain.Product{
				DateReceived: &fixedTime,
				PVZId:        &userID,
				Type:         "электроника",
			},
			inputUserRole: 1,
			mockBehavior: func(s *mock_usecase.MockPvz, product domain.Product) {
				s.EXPECT().AddProdToRecep(product).Return(domain.Product{}, errors.New("Internal Server Error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Ошибка выполнения запроса Internal Server Error"}`,
		},
		{
			name:          "Плохой ввод",
			inputUserRole: 1,
			inputBody: fmt.Sprintf(`{
				"pvzId": "%s",
				"type":1000
			}`, userID.String()),
			inputProd: domain.Product{
				DateReceived: &fixedTime,
				PVZId:        &userID,
				ReceptionId:  nil,
				Type:         "1000",
			},
			mockBehavior:         func(s *mock_usecase.MockPvz, product domain.Product) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Неверный запрос или нет активной приемки"}`,
		},
		{
			name: "Запрещен доступ",
			inputBody: fmt.Sprintf(`{
				"pvzId": "%s",
				"type":1000
			}`, userID.String()),
			inputProd: domain.Product{
				DateReceived: &fixedTime,
				PVZId:        &userID,
				ReceptionId:  nil,
				Type:         "1000",
			},
			inputUserRole: 2,
			mockBehavior: func(s *mock_usecase.MockPvz, pvz domain.Product) {
				s.EXPECT().AddProdToRecep(gomock.Any()).Times(0)
			},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Доступ запрещен"}`,
		},
		/*	{
			name: "Пустой параметр pvzId",
			inputBody: `{
				"pvzId": "invalid-uuid",
			}`,
			inputRecep: domain.ProductReception{
				DateReceived: &fixedTime,
				Status:       &stat,
				PVZId:        nil,
			},
			inputUserRole:        1,
			mockBehavior:         nil,
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Некорректный UUID ПВЗ"}`,
		},*/
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_usecase.NewMockPvz(c)

			/*var parsedPvzId uuid.UUID
			var parseErr error
			if testCase.inputBody != "" && testCase.inputBody != "invalid-uuid" {
				parsedPvzId, parseErr = uuid.Parse(testCase.inputBody)
			}

			if parseErr == nil && testCase.mockBehavior != nil {
				reception := domain.ProductReception{
					PVZId:  &parsedPvzId,
					Status: &stat,
				}
				testCase.mockBehavior(repo, reception)
			}*/
			testCase.mockBehavior(repo, testCase.inputProd)
			usecases := &usecase.Usecase{Pvz: repo}
			handler := Handler{
				Usecases: usecases,
				Now:      func() time.Time { return fixedTime },
			}
			r := gin.New()
			api := r.Group("/api")
			api.POST("/products", func(c *gin.Context) {
				c.Set("userRole", testCase.inputUserRole)
				handler.AddProducts(c)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/products", strings.NewReader(testCase.inputBody))
			req.Header.Set("Content-Type", "application/json")

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

func TestHandler_getPvz(t *testing.T) {
	type mockBehavior func(s *mock_usecase.MockPvz, params domain.GettingPvzParams)
	fixedTime := time.Date(2025, 4, 10, 15, 5, 17, 0, time.UTC)

	testTable := []struct {
		name                 string
		inputUserRole        int
		inputParams          domain.GettingPvzParams
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "OK",
			inputParams: domain.GettingPvzParams{
				Start: fixedTime,
				Page:  1,
				Limit: 10,
			},
			inputUserRole: 2,
			mockBehavior: func(s *mock_usecase.MockPvz, gettingPvz domain.GettingPvzParams) {
				s.EXPECT().GetPvz(gettingPvz).Return([]domain.PvzSummary{}, nil)
			},
			expectedStatusCode: 200,
			expectedResponseBody: `{
				"description": "Список ПВЗ",
				"content": []
			  }`,
		},
		{
			name: "Ошибка выполнения запроса",
			inputParams: domain.GettingPvzParams{
				Start: fixedTime,
				Page:  1,
				Limit: 10,
			},
			inputUserRole: 2,
			mockBehavior: func(s *mock_usecase.MockPvz, gettingPvz domain.GettingPvzParams) {
				s.EXPECT().GetPvz(gettingPvz).Return([]domain.PvzSummary{}, errors.New("Internal Server Error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Ошибка выполнения запроса Internal Server Error"}`,
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_usecase.NewMockPvz(c)
			testCase.mockBehavior(repo, testCase.inputParams)

			usecases := &usecase.Usecase{Pvz: repo}
			handler := Handler{
				Usecases: usecases,
			}
			r := gin.New()
			api := r.Group("/api")
			api.GET("/pvz", func(c *gin.Context) {
				c.Set("userRole", testCase.inputUserRole)
				handler.GetPvz(c)
			})
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/pvz?startDate=2025-04-10T15:05:17Z&page=1&limit=10", nil)

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
