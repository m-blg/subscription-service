package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // for goose
	"github.com/pressly/goose/v3"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "subscription-service/docs" // for Swagger UI
	"subscription-service/internal/config"
	"subscription-service/internal/handler"
	"subscription-service/internal/repository"
	"subscription-service/internal/service"

	swaggerFiles "github.com/swaggo/files"
)

const TIMEOUT_DEFAULT = 10 * time.Second

// @title Subscription Aggregator API
// @version 1.0
// @description REST service for managing user online subscriptions.
// @host localhost:8080
// @BasePath /api/v1
func main() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	// Config
	cfg := config.MustLoad()
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)

	// Migrations
	runMigrations(dbURL, logger)

	// pgxpool
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT_DEFAULT)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Error("unable to connect to database pool", "error", err)
		os.Exit(1)
	}
	defer func() {
		logger.Info("closing database pool")
		pool.Close()
	}()

	repo := repository.NewSubscriptionRepo(pool)
	svc := service.NewSubscriptionService(repo, logger)
	h := handler.NewSubscriptionHandler(svc)

	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// Swagger route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API Routes
	v1 := router.Group("/api/v1")
	{
		subs := v1.Group("/subscriptions")
		{
			subs.POST("/", h.Create)
			subs.GET("/", h.List)
			subs.GET("/:id", h.GetByID)
			subs.PUT("/:id", h.Update)
			subs.DELETE("/:id", h.Delete)

			subs.GET("/total", h.GetTotalCost)
		}
	}

	srv := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.Timeout,
		WriteTimeout: cfg.HTTP.Timeout,
	}

	// Start server in a separate goroutine
	go func() {
		logger.Info("service started", "port", cfg.HTTP.Port, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen and serve failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down service...")

	// Give 5 seconds to finish ongoing requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}

	logger.Info("service exited gracefully")
}

func runMigrations(dbURL string, logger *slog.Logger) {
	logger.Info("running migrations")

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		logger.Error("failed to open db for migrations", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		logger.Error("failed to set goose dialect", "error", err)
		os.Exit(1)
	}

	// Path "migrations" relative to project root
	if err := goose.Up(db, "migrations"); err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}

	logger.Info("migrations applied successfully")
}
