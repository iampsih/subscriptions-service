package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const RequestIDHeader = "X-Request-Id"

func RequestLogger(log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			req := c.Request()
			res := c.Response()

			rid := req.Header.Get(RequestIDHeader)
			if rid == "" {
				rid = res.Header().Get(RequestIDHeader)
			}

			err := next(c)

			latency := time.Since(start)

			fields := []zap.Field{
				zap.String("request_id", rid),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.Int("status", res.Status),
				zap.Duration("latency", latency),
			}

			if err != nil {
				// Echo обработает err через HTTPErrorHandler, но мы залогируем
				log.Warn("request finished with error", append(fields, zap.Error(err))...)
				return err
			}

			log.Info("request finished", fields...)
			return nil
		}
	}
}
