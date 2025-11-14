package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nedokyrill/avito-pr-api/internal/api"
	"github.com/nedokyrill/avito-pr-api/internal/server"
	"github.com/nedokyrill/avito-pr-api/internal/services/pullRequestService"
	"github.com/nedokyrill/avito-pr-api/internal/services/teamService"
	"github.com/nedokyrill/avito-pr-api/internal/services/userService"
	"github.com/nedokyrill/avito-pr-api/internal/storage/prReviewersStorage"
	"github.com/nedokyrill/avito-pr-api/internal/storage/pullRequestStorage"
	"github.com/nedokyrill/avito-pr-api/internal/storage/teamStorage"
	"github.com/nedokyrill/avito-pr-api/internal/storage/userStorage"
	"github.com/nedokyrill/avito-pr-api/pkg/consts"
	"github.com/nedokyrill/avito-pr-api/pkg/metrics"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/db"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/ginRouter"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// Init STORAGE layer
	teamRepo := teamStorage.NewTeamStorage(conn)
	userRepo := userStorage.NewUserStorage(conn)
	prRepo := pullRequestStorage.NewPullRequestStorage(conn)
	prReviewersRepo := prReviewersStorage.NewPrReviewersStorage(conn)

	// Init SERVICE layer
	teamSvc := teamService.NewTeamService(teamRepo, userRepo)
	userSvc := userService.NewUserService(userRepo, prReviewersRepo)
	prSvc := pullRequestService.NewPullRequestService(prRepo, prReviewersRepo, userRepo, teamRepo)

	// Init ROUTER
	router := ginRouter.InitRouter()

	// Register METRICS
	prometheus.MustRegister(metrics.PRLifecycleDurationHours)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Init API routes
	api.InitRoutes(router, teamSvc, userSvc, prSvc)

	// Init SERVER
	srv := server.NewAPIServer(router)

	// Start SERVER
	go srv.Start()

	// GRACEFUL SHUTDOWN
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Printf("\n")
	logger.Logger.Info("shutting down server...")
	ctx, cancel = context.WithTimeout(context.Background(), consts.GsTimeout)
	defer cancel()
	if err = srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatalw("Shutdown error",
			"error", err)
	}
}
