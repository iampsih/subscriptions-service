package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/iampsih/subscriptions-service/internal/config"
	api "github.com/iampsih/subscriptions-service/internal/http"
	"github.com/iampsih/subscriptions-service/internal/http/handlers"
	"github.com/iampsih/subscriptions-service/internal/repository/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	_ "github.com/iampsih/subscriptions-service/docs"
	echoSwagger "github.com/swaggo/echo-swagger"

	mymw "github.com/iampsih/subscriptions-service/internal/http/middleware"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	if cfg.DBDSN == "" {
		panic("DB_DSN is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	e := echo.New()

	log, _ := zap.NewProduction()
	defer log.Sync()

	e.Use(echomw.RequestID()) // добавляет X-Request-Id
	e.Use(mymw.RequestLogger(log))

	repo := postgres.NewSubscriptionRepo(pool)
	h := handlers.NewSubscriptionHandler(repo)

	api.RegisterRoutes(e, h)

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.GET("/health", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"status": "db_down",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	addr := ":" + cfg.AppPort
	e.Logger.Fatal(e.Start(addr))
	_ = os.Stdout
}
