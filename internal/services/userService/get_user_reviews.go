package userService

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/logger"
)

func (s *UserServiceImpl) GetUserReviews(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.InvalidRequest,
			"user_id query parameter is required",
		))
		return
	}

	_, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(
				domain.NotFound,
				"user not found",
			))
			return
		}
	}

	prs, err := s.prReviewersRepo.GetPRsByReviewer(ctx, userID)
	if err != nil {
		logger.Logger.Error("error getting user reviews: ", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.InternalError,
			"error getting user reviews",
		))
		return
	}

	logger.Logger.Infow("user reviews retrieved successfully", "user_id", userID)
	c.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"pull_requests": prs,
	})
}
