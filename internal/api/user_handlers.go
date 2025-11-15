package api

import (
	"github.com/gin-gonic/gin"
	"github.com/nedokyrill/avito-pr-api/internal/middleware"
	"github.com/nedokyrill/avito-pr-api/internal/services"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) InitUserHandlers(router *gin.RouterGroup) {
	usersGroup := router.Group("/users")
	{
		usersGroup.POST("/setIsActive", middleware.AuthMiddleware(), h.userService.SetIsActive)
		usersGroup.GET("/getReview", middleware.AuthMiddleware(), h.userService.GetUserReviews)
		usersGroup.POST("/deactivateTeamMembers", middleware.AuthMiddleware(), h.userService.DeactivateTeamMembers)
	}
}
