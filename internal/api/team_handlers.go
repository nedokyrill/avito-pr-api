package api

import (
	"github.com/gin-gonic/gin"
	"github.com/nedokyrill/avito-pr-api/internal/middleware"
	"github.com/nedokyrill/avito-pr-api/internal/services"
)

type TeamHandler struct {
	teamService services.TeamService
}

func NewTeamHandler(teamService services.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

func (h *TeamHandler) InitTeamHandlers(router *gin.RouterGroup) {
	teamGroup := router.Group("/team")
	{
		teamGroup.POST("/add", h.teamService.CreateTeam)
		teamGroup.GET("/get", middleware.AuthMiddleware(), h.teamService.GetTeam)
	}
}
