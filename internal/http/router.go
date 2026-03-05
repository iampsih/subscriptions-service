package http

import (
	"github.com/iampsih/subscriptions-service/internal/http/handlers"
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, h *handlers.SubscriptionHandler) {
	v1 := e.Group("/api/v1")

	v1.POST("/subscriptions", h.Create)
	v1.GET("/subscriptions", h.List)
	v1.GET("/subscriptions/total", h.Total)
	v1.GET("/subscriptions/:id", h.Get)
	v1.PUT("/subscriptions/:id", h.Update)
	v1.DELETE("/subscriptions/:id", h.Delete)
}
