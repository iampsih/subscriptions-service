package http

import (
	"github.com/iampsih/subscriptions-service/internal/http/handlers"
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, h *handlers.SubscriptionHandler) {
	v1 := e.Group("/api/v1")

	v1.POST("/subscriptions", h.Create)
}
