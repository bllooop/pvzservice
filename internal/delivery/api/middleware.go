package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeader = "Authorization"
	userCtx             = "userRole"
	userId              = "userId"
)

func (h *Handler) authIdentity(c *gin.Context) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		newErrorResponse(c, http.StatusUnauthorized, "Пустой заголовок авторизации")
		c.Abort()
		return
	}
	headerSplit := strings.Split(header, " ")
	if len(headerSplit) != 2 {
		newErrorResponse(c, http.StatusUnauthorized, "Некорректный ввод токена")
		c.Abort()
		return
	}
	if headerSplit[1] == "" {
		newErrorResponse(c, http.StatusUnauthorized, "Токен пуст")
		c.Abort()
		return
	}
	parsedId, userRole, err := h.Usecases.Authorization.ParseToken(headerSplit[1])
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		c.Abort()
		return
	}
	c.Set(userCtx, userRole)
	c.Set(userId, parsedId)
}
func getUserRole(c *gin.Context) (int, error) {
	role, ok := c.Get(userCtx)
	if !ok {
		return 0, errors.New("Роль пользователя не найдена")
	}

	roleInt, ok := role.(int)
	if !ok {
		return 0, errors.New("Роль пользователя некорректного типа данных")
	}

	return roleInt, nil
}

func getUserId(c *gin.Context) (int, error) {
	id, ok := c.Get(userId)
	if !ok {
		return 0, errors.New("ID пользователя не найдена")
	}

	idInt, ok := id.(int)
	if !ok {
		return 0, errors.New("ID пользователя некорректного типа данных")
	}

	return idInt, nil
}
