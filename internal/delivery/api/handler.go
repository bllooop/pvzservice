package api

import (
	"time"

	"github.com/bllooop/pvzservice/internal/usecase"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler struct {
	Usecases *usecase.Usecase
	Now      func() time.Time
}

func NewHandler(usecases *usecase.Usecase) *Handler {
	return &Handler{Usecases: usecases, Now: func() time.Time { return time.Now() }}
}
func NewHandlerWithFixedTime(usecases *usecase.Usecase, fixedTime time.Time) *Handler {
	return &Handler{
		Usecases: usecases,
		Now:      func() time.Time { return fixedTime },
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	router.Use(h.PrometheusMiddleware())
	router.POST("/register", h.SignUp)
	router.POST("/login", h.SignIn)
	router.POST("/dummyLogin", h.DummyLogin)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.POST("/pvz", h.authIdentity, h.CreatePvz)
	router.GET("/pvz", h.authIdentity, h.GetPvz)
	router.POST("/pvz/:pvzId/close_last_reception", h.authIdentity, h.CloseLast)
	router.POST("/pvz/:pvzId/delete_last_product", h.authIdentity, h.DeleteLast)
	router.POST("/receptions", h.authIdentity, h.CreateReceptions)
	router.POST("/products", h.authIdentity, h.AddProducts)
	return router
}
