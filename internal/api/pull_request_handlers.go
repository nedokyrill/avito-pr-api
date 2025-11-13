package api

import (
	"github.com/gin-gonic/gin"
	"github.com/nedokyrill/avito-pr-api/internal/middleware"
	"github.com/nedokyrill/avito-pr-api/internal/services"
)

type PullRequestHandler struct {
	prService services.PullRequestService
}

func NewPullRequestHandler(prService services.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{
		prService: prService,
	}
}

func (h *PullRequestHandler) InitPullRequestHandlers(router *gin.RouterGroup) {
	prGroup := router.Group("/pullRequest")
	{
		prGroup.POST("/create", middleware.AuthMiddleware(), h.prService.CreatePullRequest)
		prGroup.POST("/merge", middleware.AuthMiddleware(), h.prService.MergePullRequest)
		prGroup.POST("/reassign", middleware.AuthMiddleware(), h.prService.ReassignReviewer)
	}
}
