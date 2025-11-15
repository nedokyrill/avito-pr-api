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

// DeactivateTeamMembers выполняет массовую деактивацию пользователей команды.
// При деактивации ревьюверов автоматически переназначает их на других активных участников команды
// для всех открытых PR, где деактивируемые пользователи были назначены ревьюверами.
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

	// Нельзя деактивировать всех участников команды, некому будет ревьюить
	if len(req.UserIDs) == 0 || len(req.UserIDs) == len(team.Members) {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.InvalidRequest,
			"cannot deactivate all team members",
		))
		return
	}

	// Проверка, что все указанные пользователи действительно являются членами команды
	teamMemberIDs := make(map[string]struct{})
	for _, member := range team.Members {
		teamMemberIDs[member.UserId] = struct{}{}
	}

	for _, userID := range req.UserIDs {
		_, ok := teamMemberIDs[userID]
		if !ok {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
				domain.InvalidRequest,
				"user "+userID+" is not a member of team "+req.TeamName,
			))
			return
		}
	}

	// Поиск всех открытых PR, где деактивируемые пользователи являются ревьюверами
	prMap := s.getOpenPRsForUsers(ctx, req.UserIDs)

	// Построение плана переназначения ревьюверов
	// Для каждого открытого PR определяем, кого нужно заменить и на кого
	// Алгоритм выбирает случайных активных участников команды, исключая автора
	// Также проверяет, что после переназначения ни один PR не останется без ревьюверов
	reassignments, errResp := s.buildReassignmentsPlan(ctx, prMap, req.UserIDs, team)
	if errResp != nil {
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	// Выполнение деактивации и переназначений
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
		"reassignments_count", len(reassignments),
	)

	response := domain.DeactivateTeamMembersResponse{
		DeactivatedUserIDs: deactivatedUserIDs,
		Reassignments:      reassignments,
	}

	c.JSON(http.StatusOK, response)
}

func (s *UserServiceImpl) getOpenPRsForUsers(ctx context.Context, userIDs []string) map[string]domain.PullRequestShort {
	prMap := make(map[string]domain.PullRequestShort)

	// Для каждого деактивируемого пользователя получаем список PR, где он ревьювер
	for _, userID := range userIDs {
		prs, err := s.prReviewersRepo.GetPRsByReviewer(ctx, userID)
		if err != nil {
			logger.Logger.Errorw("error getting PRs for reviewer",
				"user_id", userID,
				"error", err)
			continue
		}

		// Фильтруем только открытые PR (статус OPEN)
		// Мердженные PR не трогаем - для них нельзя менять ревьюверов
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
) ([]domain.ReviewerReassignment, *domain.ErrorResponse) {
	var reassignments []domain.ReviewerReassignment

	// Множество деактивируемых пользователей
	usersToDeactivateSet := make(map[string]struct{})
	for _, userID := range usersToDeactivate {
		usersToDeactivateSet[userID] = struct{}{}
	}

	// Оставляем только тех, кто останется активным после деактивации - эти пользователи могут быть
	// назначены новыми ревьюверами
	availableMembers := make([]domain.TeamMember, 0)
	for _, member := range team.Members {
		_, ok := usersToDeactivateSet[member.UserId]
		if !ok {
			availableMembers = append(availableMembers, member)
		}
	}

	for _, pr := range prMap {
		currentReviewers, err := s.prReviewersRepo.GetAssignedReviewers(ctx, pr.PullRequestId)
		if err != nil {
			logger.Logger.Errorw("error getting reviewers for PR",
				"pr_id", pr.PullRequestId,
				"error", err)
			continue
		}

		// Собираем ревьюверов, которых нужно заменить
		reviewersToReplace := make([]string, 0)

		// Создаем множество уже назначенных ревьюверов, чтобы не назначать одного и того же пользователя дважды
		alreadyAssigned := make(map[string]struct{})
		for _, r := range currentReviewers {
			alreadyAssigned[r] = struct{}{}
			_, ok := usersToDeactivateSet[r]
			if ok {
				reviewersToReplace = append(reviewersToReplace, r)
			}
		}

		// Если нечего заменять (все ревьюверы остаются активными), пропускаем этот пр
		if len(reviewersToReplace) == 0 {
			continue
		}

		// здесь используем не заданное число ревьюеров (2), а столько, сколько их уже было
		candidates := utils.RandSelectReviewers(availableMembers, pr.AuthorId, len(reviewersToReplace))

		// Фильтруем кандидатов: исключаем уже назначенных ревьюверов
		availableCandidates := make([]string, 0)
		for _, candidate := range candidates {
			_, ok := alreadyAssigned[candidate]
			if !ok {
				availableCandidates = append(availableCandidates, candidate)
			}
		}

		// Распределяем кандидатов по ревьюверам последовательно
		candidateIndex := 0
		addedCount := 0 // Считаем, сколько ревьюверов реально будет назначено
		for _, reviewerID := range reviewersToReplace {
			var newReviewerID string
			// Если есть доступные кандидаты, назначаем их по очереди
			if candidateIndex < len(availableCandidates) {
				newReviewerID = availableCandidates[candidateIndex]
				// Помечаем кандидата как уже назначенного, чтобы не использовать его повторно
				alreadyAssigned[newReviewerID] = struct{}{}
				candidateIndex++
				addedCount++
			}
			// Если кандидатов нет, newReviewerID останется пустым
			// Это означает, что ревьювер будет удален без замены

			reassignments = append(reassignments, domain.ReviewerReassignment{
				PrID:          pr.PullRequestId,
				OldReviewerID: reviewerID,
				NewReviewerID: newReviewerID,
			})
		}

		// Проверяем, что после переназначения пр не останется без ревьюверов
		finalReviewerCount := len(currentReviewers) - len(reviewersToReplace) + addedCount

		if finalReviewerCount == 0 {
			errRes := domain.NewErrorResponse(
				domain.NoCandidate,
				"cannot deactivate reviewers: PR "+pr.PullRequestId+" would be left without reviewers",
			)
			return nil, &errRes
		}
	}

	return reassignments, nil
}
