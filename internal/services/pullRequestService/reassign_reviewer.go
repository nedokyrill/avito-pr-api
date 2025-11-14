package pullRequestService

import (
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/pkg/utils"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/logger"
)

func (s *PullRequestServiceImpl) ReassignReviewer(c *gin.Context) {
	ctx := c.Request.Context()

	var req domain.ReassignReviewerRequest

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
			domain.ErrReassignReviewerMsg,
		))
		return
	}

	if pr.Status == domain.PullRequestStatusMERGED {
		c.JSON(http.StatusConflict, domain.NewErrorResponse(
			domain.PrMerged,
			"cannot reassign on merged PR",
		))
		return
	}

	assignedReviewers, err := s.prReviewersRepo.GetAssignedReviewers(ctx, req.PullRequestID)
	if err != nil {
		logger.Logger.Error("error getting assigned reviewers: ", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.InternalError,
			domain.ErrReassignReviewerMsg,
		))
		return
	}

	if !utils.Contains(assignedReviewers, req.OldUserID) {
		c.JSON(http.StatusConflict, domain.NewErrorResponse(
			domain.NotAssigned,
			"reviewer is not assigned to this PR",
		))
		return
	}

	oldUser, err := s.userRepo.GetUserByID(ctx, req.OldUserID)
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
			domain.ErrReassignReviewerMsg,
		))
		return
	}

	team, err := s.teamRepo.GetTeamByName(ctx, oldUser.TeamName)
	if err != nil {
		logger.Logger.Error("error getting team: ", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.InternalError,
			domain.ErrReassignReviewerMsg,
		))
		return
	}

	var candidates []string
	for _, member := range team.Members {
		if member.IsActive && member.UserId != pr.AuthorId && !utils.Contains(assignedReviewers, member.UserId) {
			candidates = append(candidates, member.UserId)
		}
	}

	if len(candidates) == 0 {
		err = s.prRepo.SetNeedMoreReviewers(ctx, req.PullRequestID, true)
		if err != nil {
			logger.Logger.Error("error setting need_more_reviewers flag: ", err)
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
				domain.InternalError,
				domain.ErrReassignReviewerMsg,
			))
			return
		}
	}

	//nolint:gosec // G404: math/rand достаточно для случайного выбора ревьюеров
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	newReviewerID := candidates[rng.Intn(len(candidates))]

	err = s.prReviewersRepo.ReassignReviewerAtomic(ctx, req.PullRequestID, req.OldUserID, newReviewerID)
	if err != nil {
		logger.Logger.Error("error reassigning reviewer: ", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.InternalError,
			domain.ErrReassignReviewerMsg,
		))
		return
	}

	updatedReviewers, err := s.prReviewersRepo.GetAssignedReviewers(ctx, req.PullRequestID)
	if err != nil {
		logger.Logger.Error("error getting updated reviewers: ", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.InternalError,
			domain.ErrReassignReviewerMsg,
		))
		return
	}

	pr.AssignedReviewers = updatedReviewers

	logger.Logger.Infow("reviewer reassigned successfully",
		"pr_id", req.PullRequestID,
		"old_user_id", req.OldUserID,
		"new_user_id", newReviewerID,
	)
	c.JSON(http.StatusOK, gin.H{
		"pr":          pr,
		"replaced_by": newReviewerID,
	})
}
