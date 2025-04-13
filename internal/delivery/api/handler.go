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
	return &Handler{Usecases: usecases, Now: func() time.Time { return time.Date(2025, 4, 10, 15, 5, 17, 329922000, time.UTC) }}
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
	api := router.Group("/api", h.authIdentity)
	{
		api.POST("/pvz", h.CreatePvz)
		api.GET("/pvz", h.GetPvz)
		api.POST("/pvz/:pvzId/close_last_reception", h.CloseLast)
		api.POST("/pvz/:pvzId/delete_last_product", h.DeleteLast)
		api.POST("/receptions", h.CreateReceptions)
		api.POST("/products", h.AddProducts)
	}
	return router
}
