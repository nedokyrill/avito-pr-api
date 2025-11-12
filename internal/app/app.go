package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nedokyrill/avito-pr-api/internal/server"
	"github.com/nedokyrill/avito-pr-api/pkg/consts"
	"github.com/nedokyrill/avito-pr-api/pkg/db"
	"github.com/nedokyrill/avito-pr-api/pkg/logger"
)

func Run() {
	// Init LOGGER
	logger.InitLogger()

	// Load ENVIRONMENT VARIABLES
	err := godotenv.Load()
	if err != nil {
		logger.Logger.Fatal("error loading .env file, exiting...")
	}

	// Connect to DATABASE
	ctx, cancel := context.WithTimeout(context.Background(), consts.PgxTimeout)
	defer cancel()

	conn, err := db.Connect(ctx)
	if err != nil {
		logger.Logger.Fatal("error connecting to database, exiting...")
	}
	defer conn.Close()

	// Init ROUTER
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Init SERVER
	srv := server.NewAPIServer(router)

	// Start SERVER
	go srv.Start()

	// GRACEFUL SHUTDOWN
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Logger.Info("Shutting down server...")
	ctx, cancel = context.WithTimeout(context.Background(), consts.GsTimeout)
	defer cancel()
	if err = srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatalw("Shutdown error",
			"error", err)
	}
}
