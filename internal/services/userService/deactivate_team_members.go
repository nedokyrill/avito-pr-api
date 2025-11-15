package userService

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/pkg/utils"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/logger"
)

func (s *UserServiceImpl) DeactivateTeamMembers(c *gin.Context) {
	ctx := c.Request.Context()

	var req domain.DeactivateTeamMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.InvalidRequest,
			"invalid request body",
		))
		return
	}

	// Получаем и валидируем команду
	team, err := s.teamRepo.GetTeamByName(ctx, req.TeamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(
				domain.NotFound,
				"team not found",
			))
			return
		}
		logger.Logger.Error("error getting team: ", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.InternalError,
			domain.ErrDeactivatingUsersMsg,
		))
		return
	}

	// Валидация запроса на деактивацию
	if len(req.UserIDs) == 0 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.InvalidRequest,
			"cannot deactivate all team members",
		))
		return
	}

	if len(req.UserIDs) == len(team.Members) {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.InvalidRequest,
			"cannot deactivate all team members",
		))
		return
	}

	teamMemberIDs := make(map[string]bool)
	for _, member := range team.Members {
		teamMemberIDs[member.UserId] = true
	}

	for _, userID := range req.UserIDs {
		if !teamMemberIDs[userID] {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
				domain.InvalidRequest,
				"user "+userID+" is not a member of team "+req.TeamName,
			))
			return
		}
	}

	prMap := s.getOpenPRsForUsers(ctx, req.UserIDs)
	reassignments := s.buildReassignmentsPlan(ctx, prMap, req.UserIDs, team)

	deactivatedUserIDs, err := s.teamRepo.DeactivateTeamMembers(ctx, req.TeamName, req.UserIDs, reassignments)
	if err != nil {
		logger.Logger.Error("error deactivating users atomically: ", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.InternalError,
			domain.ErrDeactivatingUsersMsg,
		))
		return
	}

	logger.Logger.Infow("team members deactivated",
		"team_name", req.TeamName,
		"deactivated_count", len(deactivatedUserIDs),
		"reassignments_count", len(reassignments))

	response := domain.DeactivateTeamMembersResponse{
		DeactivatedUserIDs: deactivatedUserIDs,
		Reassignments:      reassignments,
	}

	c.JSON(http.StatusOK, response)
}

func (s *UserServiceImpl) getOpenPRsForUsers(ctx context.Context, userIDs []string) map[string]domain.PullRequestShort {
	prMap := make(map[string]domain.PullRequestShort)
	for _, userID := range userIDs {
		prs, err := s.prReviewersRepo.GetPRsByReviewer(ctx, userID)
		if err != nil {
			logger.Logger.Errorw("error getting PRs for reviewer",
				"user_id", userID,
				"error", err)
			continue
		}

		for _, pr := range prs {
			if pr.Status == domain.PullRequestStatusOPEN {
				prMap[pr.PullRequestId] = pr
			}
		}
	}
	return prMap
}

func (s *UserServiceImpl) buildReassignmentsPlan(
	ctx context.Context,
	prMap map[string]domain.PullRequestShort,
	usersToDeactivate []string,
	team *domain.Team,
) []domain.ReviewerReassignment {
	var reassignments []domain.ReviewerReassignment

	for _, pr := range prMap {
		currentReviewers, err := s.prReviewersRepo.GetAssignedReviewers(ctx, pr.PullRequestId)
		if err != nil {
			logger.Logger.Errorw("error getting reviewers for PR",
				"pr_id", pr.PullRequestId,
				"error", err)
			continue
		}

		alreadyAssigned := make(map[string]bool)
		for _, r := range currentReviewers {
			alreadyAssigned[r] = true
		}

		for _, reviewerID := range currentReviewers {
			if !utils.Contains(usersToDeactivate, reviewerID) {
				continue
			}

			// Выбираем случайного активного члена команды (исключая автора)
			candidates := utils.RandSelectReviewers(team.Members, pr.AuthorId, 1)
			var newReviewerID string
			for _, candidate := range candidates {
				if !alreadyAssigned[candidate] {
					newReviewerID = candidate
					alreadyAssigned[candidate] = true
					break
				}
			}

			reassignments = append(reassignments, domain.ReviewerReassignment{
				PrID:          pr.PullRequestId,
				OldReviewerID: reviewerID,
				NewReviewerID: newReviewerID,
			})
		}
	}

	return reassignments
}
