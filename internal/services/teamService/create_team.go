package teamService

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/logger"
)

func (s *TeamServiceImpl) CreateTeam(c *gin.Context) {
	ctx := c.Request.Context()

	var team domain.Team

	if err := c.ShouldBindJSON(&team); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.InvalidRequest,
			"invalid request body",
		))
		return
	}

	teamID, err := s.teamRepo.CreateTeam(ctx, team.TeamName)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
				domain.TeamExists,
				"team_name already exists",
			))
			return
		}
		logger.Logger.Error("error creating team: ", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
			domain.InternalError,
			"error creating team",
		))
		return
	}

	for i := range team.Members {
		member := &team.Members[i]
		user := &domain.User{
			UserId:   member.UserId,
			Username: member.Username,
			TeamName: team.TeamName,
			IsActive: member.IsActive,
		}

		err = s.userRepo.CreateOrUpdateUser(ctx, user, teamID)
		if err != nil {
			logger.Logger.Error("error creating/updating user: ", err)
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(
				domain.InternalError,
				"error creating/updating user",
			))
			return
		}
	}

	logger.Logger.Infow("team created successfully", "team_name", team.TeamName)
	c.JSON(http.StatusCreated, gin.H{
		"team": team,
	})
}
