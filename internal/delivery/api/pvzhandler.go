package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bllooop/pvzservice/internal/domain"
	logger "github.com/bllooop/pvzservice/pkg/logging"
	prometheus "github.com/bllooop/pvzservice/prometheus"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) CreatePvz(c *gin.Context) {

	logger.Log.Info().Msg("Получен запрос на заведение ПВЗ")
	if c.Request.Method != http.MethodPost {
		logger.Log.Error().Msg("Требуется запрос POST")
		newErrorResponse(c, http.StatusBadRequest, "Неверный запрос")
		return
	}
	userRole, err := getUserRole(c)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, "Ошибка получения роли "+err.Error())
		return
	}
	logger.Log.Debug().Msgf("Успешно получена роль %v", getRoleName(userRole))
	if userRole != 2 {
		logger.Log.Error().Msg("Данный запрос доступен только модератору")
		newErrorResponse(c, http.StatusBadRequest, "Доступ запрещен")
		return
	}
	var input domain.PVZ
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Log.Error().Err(err).Msg(err.Error())
		newErrorResponse(c, http.StatusBadRequest, "Неверный запрос")
		return
	}
	logger.Log.Debug().Msgf("Успешно прочитаны данные из запроса  %s", input.City)
	now := h.Now()
	input.DateRegister = &now
	result, err := h.Usecases.Pvz.CreatePvz(input)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, "Ошибка выполнения запроса "+err.Error())
		return
	}
	prometheus.NumOfCreatedPVZ.Inc()
	c.JSON(http.StatusOK, map[string]any{
		"message": "ПВЗ создан",
		"data":    result,
	})
	logger.Log.Info().Msg("Получен ответ cоздание пвз")
}

func (h *Handler) GetPvz(c *gin.Context) {
	logger.Log.Info().Msg("Получен запрос на получение данны о ПВЗ")
	if c.Request.Method != http.MethodGet {
		logger.Log.Error().Msg("Требуется запрос GET")
		newErrorResponse(c, http.StatusBadRequest, "Требуется запрос GET")
		return
	}
	var startParse, endParse time.Time
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 {
		limitInt = 10
	} else if limitInt > 30 {
		limitInt = 30
	}
	if startDate != "" {
		startParse, err = time.Parse(time.RFC3339, startDate)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			return
		}
	}
	if endDate != "" {
		endParse, err = time.Parse(time.RFC3339, endDate)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			return
		}
	}
	input := domain.GettingPvzParams{
		Start: startParse,
		End:   endParse,
		Page:  pageInt,
		Limit: limitInt,
	}
	logger.Log.Debug().Msgf("Успешно прочитаны параметры из запроса %s, %s,%v,%v", startParse, endParse, pageInt, limitInt)
	result, err := h.Usecases.GetPvz(input)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, "Ошибка выполнения запроса "+err.Error())
		return
	}
	r := result[0]

	logger.Log.Info().Msg("Получен ответ на запрос информации о ПВЗ")
	/*c.JSON(http.StatusOK, map[string]any{
		"description": "Список ПВЗ",
		"content":     result,
	})*/
	c.JSON(http.StatusOK, map[string]any{
		"description": "Список ПВЗ",
		"pvz":         r.PvzInfo,
		"receptions":  r.ReceptionsInfo,
	})
}

func (h *Handler) CloseLast(c *gin.Context) {
	logger.Log.Info().Msg("Получен запрос на закрытие приёмки")
	if c.Request.Method != http.MethodPost {
		logger.Log.Error().Msg("Требуется запрос POST")
		newErrorResponse(c, http.StatusBadRequest, "Требуется запрос POST")
		return
	}
	pvzIdParam := c.Param("pvzId")

	pvzId, err := uuid.Parse(pvzIdParam)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "Некорректный UUID ПВЗ")
		return
	}
	logger.Log.Debug().Msgf("Успешно прочитан параметр из запроса %s", pvzId)
	userRole, err := getUserRole(c)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, "Ошибка получения роли "+err.Error())
		return
	}
	logger.Log.Debug().Msgf("Успешно получена роль %v", userRole)
	if userRole != 1 {
		logger.Log.Error().Msg("Данный запрос доступен только сотруднику ПВЗ")
		newErrorResponse(c, http.StatusBadRequest, "Доступ запрещен")
		return
	}
	result, err := h.Usecases.Pvz.CloseReception(pvzId)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, "Ошибка выполнения запроса "+err.Error())
		return
	}
	logger.Log.Info().Msg("Получен ответ на закрытие приемки")
	c.JSON(http.StatusOK, map[string]any{
		"message": "Приемка закрыта",
		"data":    result,
	})
}
func (h *Handler) DeleteLast(c *gin.Context) {
	logger.Log.Info().Msg("Получен запрос на удаление товаров в рамках не закрытой приёмки:")
	if c.Request.Method != http.MethodPost {
		logger.Log.Error().Msg("Требуется запрос POST")
		newErrorResponse(c, http.StatusBadRequest, "Требуется запрос POST")
		return
	}
	pvzIdParam := c.Param("pvzId")

	pvzId, err := uuid.Parse(pvzIdParam)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "Некорректный UUID ПВЗ")
		return
	}
	logger.Log.Debug().Msgf("Успешно прочитан параметр из запроса %s", pvzId)
	userRole, err := getUserRole(c)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, "Ошибка получения роли "+err.Error())
		return
	}
	logger.Log.Debug().Msgf("Успешно получена роль %v", userRole)
	if userRole != 1 {
		logger.Log.Error().Msg("Данный запрос доступен только сотруднику ПВЗ")
		newErrorResponse(c, http.StatusBadRequest, "Доступ запрещен")
		return
	}
	err = h.Usecases.Pvz.DeleteLastProduct(pvzId)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, "Ошибка выполнения запроса "+err.Error())
		return
	}
	logger.Log.Info().Msg("Получен ответ на удаление товара")
	c.JSON(http.StatusOK, map[string]any{
		"message": "Товар удален",
	})
}
func (h *Handler) CreateReceptions(c *gin.Context) {
	logger.Log.Info().Msg("Получен запрос на добавление информации о приёмке товаров")
	if c.Request.Method != http.MethodPost {
		logger.Log.Error().Msg("Требуется запрос POST")
		newErrorResponse(c, http.StatusBadRequest, "Требуется запрос POST")
		return
	}
	userRole, err := getUserRole(c)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, "Ошибка получения роли "+err.Error())
		return
	}
	logger.Log.Debug().Msgf("Успешно получена роль %v", userRole)
	if userRole != 1 {
		logger.Log.Error().Msg("Данный запрос доступен только сотруднику ПВЗ")
		newErrorResponse(c, http.StatusBadRequest, "Доступ запрещен")
		return
	}
	var input domain.ProductReception
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Log.Error().Err(err).Msg(err.Error())
		newErrorResponse(c, http.StatusBadRequest, "Неверный запрос или есть незакрытая приемка")
		return
	}
	if input.PVZId == nil || *input.PVZId == uuid.Nil {
		logger.Log.Error().Msg("Некорректный UUID ПВЗ")
		newErrorResponse(c, http.StatusBadRequest, "Некорректный UUID ПВЗ")
		return
	}
	logger.Log.Debug().Msgf("Успешно прочитаны данные из запроса %s", input.PVZId)
	now := h.Now()
	input.DateReceived = &now
	status := "in_progress"
	input.Status = &status
	result, err := h.Usecases.Pvz.CreateRecep(input)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, "Ошибка выполнения запроса "+err.Error())
		return
	}
	prometheus.NumOfCreatedRecep.Inc()

	logger.Log.Info().Msg("Получен ответ на добавление информации о приемке")
	c.JSON(http.StatusOK, map[string]any{
		"message": "Приемка создана",
		"data":    result,
	})

}

func (h *Handler) AddProducts(c *gin.Context) {
	logger.Log.Info().Msg("Получен запрос на добавление товаров в рамках одной приёмки")
	if c.Request.Method != http.MethodPost {
		logger.Log.Error().Msg("Требуется запрос POST")
		newErrorResponse(c, http.StatusBadRequest, "Требуется запрос POST")
		return
	}
	userRole, err := getUserRole(c)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, "Ошибка получения роли "+err.Error())
		return
	}
	logger.Log.Debug().Msgf("Успешно получена роль %v", userRole)
	if userRole != 1 {
		logger.Log.Error().Msg("Данный запрос доступен только сотруднику ПВЗ")
		newErrorResponse(c, http.StatusBadRequest, "Доступ запрещен")
		return
	}
	var input domain.Product
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Log.Error().Err(err).Msg(err.Error())
		newErrorResponse(c, http.StatusBadRequest, "Неверный запрос или нет активной приемки")
		return
	}
	if *input.PVZId == uuid.Nil {
		logger.Log.Error().Err(err).Msg(err.Error())
		newErrorResponse(c, http.StatusBadRequest, "Неверный запрос или нет активной приемки")
		return
	}

	logger.Log.Debug().Msgf("Успешно прочитаны данные из запроса %s, %s", input.Type, input.PVZId)
	now := h.Now()
	input.DateReceived = &now
	result, err := h.Usecases.Pvz.AddProdToRecep(input)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, "Ошибка выполнения запроса "+err.Error())
		return
	}

	prometheus.NumOfAddedProducts.Inc()
	logger.Log.Info().Msg("Получен ответ на добавление товаров")
	c.JSON(http.StatusOK, map[string]any{
		"message": "Товар добавлен",
		"data":    result,
	})
}

func getRoleName(userRole int) string {
	for k, v := range roleMap {
		if v == userRole {
			return k
		}
	}

	return "Unknown role"
}
