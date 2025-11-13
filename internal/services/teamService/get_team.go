package teamService

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/internal/storage/teamStorage"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/logger"
)

func (s *TeamServiceImpl) GetTeam(c *gin.Context) {
	ctx := c.Request.Context()

	teamName := c.Query("team_name")
	if teamName == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(
			domain.InvalidRequest,
			"team_name query parameter is required",
		))
		return
	}

	team, err := s.teamRepo.GetTeamByName(ctx, teamName)
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

	logger.Logger.Infow("team retrieved successfully", "team_name", teamName)
	c.JSON(http.StatusOK, team)
}
