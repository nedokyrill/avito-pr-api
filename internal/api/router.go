package api

import (
	"github.com/gin-gonic/gin"
	"github.com/nedokyrill/avito-pr-api/internal/services"
)

func InitRoutes(
	router *gin.Engine,
	teamService services.TeamService,
	userService services.UserService,
	prService services.PullRequestService,
) {
	teamHandler := NewTeamHandler(teamService)
	userHandler := NewUserHandler(userService)
	prHandler := NewPullRequestHandler(prService)

	api := router.Group("/")

	teamHandler.InitTeamHandlers(api)
	userHandler.InitUserHandlers(api)
	prHandler.InitPullRequestHandlers(api)
}
