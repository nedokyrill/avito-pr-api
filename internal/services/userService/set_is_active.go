package userService

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/logger"
)

func (s *UserServiceImpl) SetIsActive(c *gin.Context) {
	ctx := c.Request.Context()

	var req domain.SetIsActiveRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.InvalidRequest,
			"invalid request body",
		))
		return
	}

	err := s.userRepo.SetUserIsActive(ctx, req.UserID, req.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(
				domain.NotFound,
				"user not found",
			))
			return
		}
		logger.Logger.Error("error setting user active status: ", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.InternalError,
			"error setting user active status",
		))
		return
	}

	user, err := s.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(
				domain.NotFound,
				"user not found",
			))
			return
		}
		logger.Logger.Error("error getting user: ", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.InternalError,
			"error getting user",
		))
		return
	}

	logger.Logger.Infow("user active status updated", "user_id", req.UserID, "is_active", req.IsActive)
	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
