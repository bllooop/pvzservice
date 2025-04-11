package api

import (
	"net/http"

	"github.com/bllooop/pvzservice/internal/domain"
	logger "github.com/bllooop/pvzservice/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var roleMap = map[string]int{
	"employee":  1,
	"moderator": 2,
}

var dummyUserIds = map[string]string{
	"employee":  "11111111-1111-1111-1111-111111111111",
	"moderator": "22222222-2222-2222-2222-222222222222",
}

func (h *Handler) DummyLogin(c *gin.Context) {
	logger.Log.Info().Msg("Получили запрос на получение токена")
	if c.Request.Method != http.MethodPost {
		newErrorResponse(c, http.StatusBadRequest, "Требуется запрос POST")
		logger.Log.Error().Msg("Требуется запрос POST")
		return
	}
	var input domain.DummyLogin
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Log.Error().Err(err).Msg(err.Error())
		newErrorResponse(c, http.StatusBadRequest, "Неверный запрос")
		return
	}
	logger.Log.Debug().Msgf("Успешно прочитана роль: %s", input.Role)
	userRole, ok := roleMap[input.Role]
	if !ok {
		newErrorResponse(c, http.StatusBadRequest, "Неверная роль")
		return
	}

	userIdStr, ok := dummyUserIds[input.Role]
	if !ok {
		newErrorResponse(c, http.StatusBadRequest, "UUID не задан для роли")
		return
	}
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, "Ошибка UUID: "+err.Error())
		logger.Log.Error().Err(err).Msg("Невалидный UUID")
		return
	}
	logger.Log.Debug().Any("AAAAA", userRole)

	token, err := h.Usecases.Authorization.GenerateToken(userId, userRole)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, "Ошибка создания токена: "+err.Error())
		logger.Log.Error().Err(err).Msg("")
		return
	}

	c.JSON(http.StatusOK, map[string]any{
		"token": token,
	})
	logger.Log.Info().Msg("Получили токен")
}
func (h *Handler) SignUp(c *gin.Context) {
	logger.Log.Info().Msg("Получили запрос на создание пользователя")
	if c.Request.Method != http.MethodPost {
		newErrorResponse(c, http.StatusBadRequest, "Требуется запрос POST")
		logger.Log.Error().Msg("Требуется запрос POST")
		return
	}
	var input domain.User
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Log.Error().Err(err).Msg(err.Error())
		newErrorResponse(c, http.StatusBadRequest, "Неверный запрос")
		return
	}
	logger.Log.Debug().Msgf("Успешно прочитаны почта: %s, пароль: %s, роль: %s", input.Email, input.Password, input.Role)
	id, err := h.Usecases.Authorization.CreateUser(input)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})
	logger.Log.Info().Msg("Создали пользователя")

}

func (h *Handler) SignIn(c *gin.Context) {
	logger.Log.Info().Msg("Получили запрос на авторизацию пользователя")
	var input domain.SignInInput
	if c.Request.Method != http.MethodPost {
		logger.Log.Error().Msg("Требуется запрос POST")
		newErrorResponse(c, http.StatusBadRequest, "Требуется запрос POST")
		return
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Log.Error().Err(err).Msg(err.Error())
		newErrorResponse(c, http.StatusBadRequest, "Неверный запрос")
		return
	}
	logger.Log.Debug().Msgf("Успешно прочитаны почта: %s, пароль: %s", input.Email, input.Password)
	user, err := h.Usecases.Authorization.SignUser(input.Email, input.Password)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, "Ошибка авторизации: "+err.Error())
		return
	}
	userRole, ok := roleMap[user.Role]
	if !ok {
		newErrorResponse(c, http.StatusBadRequest, "Неверная роль")
		return
	}
	token, err := h.Usecases.Authorization.GenerateToken(user.Id, userRole)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, "Ошибка создания токена: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
	})
	logger.Log.Info().Msg("Получили токен")
}
