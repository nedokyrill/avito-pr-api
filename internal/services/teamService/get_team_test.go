package teamService

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/internal/storage/mocks"
	"github.com/nedokyrill/avito-pr-api/internal/storage/teamStorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamService_GetTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamRepo := mocks.NewMockTeamRepositoryInterface(ctrl)
	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	service := NewTeamService(mockTeamRepo, mockUserRepo)

	t.Run("successfully get team", func(t *testing.T) {
		team := &domain.Team{
			TeamName: "Backend Team",
			Members: []domain.TeamMember{
				{UserId: uuid.NewString(), Username: "Alice", IsActive: true},
				{UserId: uuid.NewString(), Username: "Bob", IsActive: true},
			},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/teams?team_name=Backend%20Team", nil)

		mockTeamRepo.EXPECT().
			GetTeamByName(gomock.Any(), "Backend Team").
			Return(team, nil)

		service.GetTeam(c)

		require.Equal(t, http.StatusOK, w.Code)
		var response domain.Team
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Backend Team", response.TeamName)
		assert.Len(t, response.Members, 2)
	})

	t.Run("missing team_name parameter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/teams", nil)

		service.GetTeam(c)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var response domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, domain.InvalidRequest, response.Error.Code)
	})

	t.Run("team not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/teams?team_name=NonExistent", nil)

		mockTeamRepo.EXPECT().
			GetTeamByName(gomock.Any(), "NonExistent").
			Return(nil, teamStorage.ErrTeamNotExists)

		service.GetTeam(c)

		require.Equal(t, http.StatusNotFound, w.Code)
		var response domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, domain.NotFound, response.Error.Code)
	})

	t.Run("internal server error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/teams?team_name=Backend%20Team", nil)

		mockTeamRepo.EXPECT().
			GetTeamByName(gomock.Any(), "Backend Team").
			Return(nil, errors.New("database connection error"))

		service.GetTeam(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		var response domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, domain.InternalError, response.Error.Code)
	})
}

