package pullRequestService

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/internal/storage/teamStorage"
	"github.com/nedokyrill/avito-pr-api/pkg/utils"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/logger"
)

func (s *PullRequestServiceImpl) CreatePullRequest(c *gin.Context) {
	ctx := c.Request.Context()

	var req domain.CreatePullRequestRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.InvalidRequest,
			"invalid request body",
		))
		return
	}

	author, err := s.userRepo.GetUserByID(ctx, req.AuthorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(
				domain.NotFound,
				"author not found",
			))
			return
		}
		logger.Logger.Error("error getting author: ", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.InternalError,
			"error getting author",
		))
		return
	}

	team, err := s.teamRepo.GetTeamByName(ctx, author.TeamName)
	if err != nil {
		if errors.Is(err, teamStorage.ErrTeamNotExists) {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(
				domain.NotFound,
				"team not found",
			))
			return
		}
		logger.Logger.Error("error getting team: ", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.InternalError,
			"error getting team",
		))
		return
	}

	now := time.Now()
	pr := &domain.PullRequest{
		PullRequestId:     req.PullRequestID,
		PullRequestName:   req.PullRequestName,
		AuthorId:          req.AuthorID,
		Status:            domain.PullRequestStatusOPEN,
		AssignedReviewers: []string{},
		CreatedAt:         &now,
		MergedAt:          nil,
	}

	err = s.prRepo.CreatePullRequest(ctx, pr)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, domain.NewErrorResponse(
				domain.PrExists,
				"PR id already exists",
			))
			return
		}
		logger.Logger.Error("error creating PR: ", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.InternalError,
			"error creating pull request",
		))
		return
	}

	reviewers := utils.RandSelectReviewers(team.Members, req.AuthorID, domain.MaxReviewersCount)

	for _, reviewerID := range reviewers {
		err = s.prReviewersRepo.AddReviewer(ctx, req.PullRequestID, reviewerID)
		if err != nil {
			logger.Logger.Error("error adding reviewer: ", err)
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
				domain.InternalError,
				"error adding reviewer",
			))
			return
		}
	}

	pr.AssignedReviewers = reviewers

	needMore := len(reviewers) < domain.MaxReviewersCount
	pr.NeedMoreReviewers = &needMore
	if needMore {
		err = s.prRepo.SetNeedMoreReviewers(ctx, req.PullRequestID, true)
		if err != nil {
			logger.Logger.Error("error setting need_more_reviewers flag: ", err)
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
				domain.InternalError,
				"error setting need_more_reviewers flag",
			))
			return
		}
	}

	logger.Logger.Infow("PR created successfully", "pr_id", req.PullRequestID, "reviewers_count", len(reviewers))
	c.JSON(http.StatusCreated, gin.H{
		"pr": pr,
	})
}
