package pullRequestService

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/logger"
)

func (s *PullRequestServiceImpl) MergePullRequest(c *gin.Context) {
	ctx := c.Request.Context()

	var req domain.MergePullRequestRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.InvalidRequest,
			"invalid request body",
		))
		return
	}

	pr, err := s.prRepo.GetPullRequestByID(ctx, req.PullRequestID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(
				domain.NotFound,
				"PR not found",
			))
			return
		}
		logger.Logger.Error("error getting PR: ", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.InternalError,
			"error getting pull request",
		))
		return
	}

	if pr.Status == domain.PullRequestStatusOPEN {
		err = s.prRepo.MergePullRequest(ctx, req.PullRequestID)
		if err != nil {
			logger.Logger.Error("error merging PR: ", err)
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
				domain.InternalError,
				"error merging pull request",
			))
			return
		}
		
		now := time.Now()
		pr.MergedAt = &now
		pr.Status = domain.PullRequestStatusMERGED
	}

	reviewers, err := s.prReviewersRepo.GetAssignedReviewers(ctx, req.PullRequestID)
	if err != nil {
		logger.Logger.Error("error getting reviewers: ", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.InternalError,
			"error getting reviewers",
		))
		return
	}
	pr.AssignedReviewers = reviewers

	logger.Logger.Infow("PR merged successfully", "pr_id", req.PullRequestID)
	c.JSON(http.StatusOK, gin.H{
		"pr": pr,
	})
}
