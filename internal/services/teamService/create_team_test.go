package teamService

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/golang/mock/gomock"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/internal/storage/mocks"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
	logger.Logger = zap.NewNop().Sugar()
}

func TestTeamService_CreateTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamRepo := mocks.NewMockTeamRepositoryInterface(ctrl)
	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	service := NewTeamService(mockTeamRepo, mockUserRepo)

	t.Run("successfully create team with members", func(t *testing.T) {
		teamID := uuid.New()
		user1ID := "user-alice-1"
		user2ID := "user-bob-1"

		requestBody := `{
			"team_name": "Backend Team",
			"members": [
				{"user_id": "` + user1ID + `", "username": "Alice", "is_active": true},
				{"user_id": "` + user2ID + `", "username": "Bob", "is_active": true}
			]
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/teams", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockTeamRepo.EXPECT().
			CreateTeamWithMembers(gomock.Any(), "Backend Team", gomock.Any()).
			Return(teamID, nil)

		service.CreateTeam(c)

		require.Equal(t, http.StatusCreated, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "team")
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/teams", strings.NewReader("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		service.CreateTeam(c)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var response domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, domain.InvalidRequest, response.Error.Code)
	})

	t.Run("team already exists", func(t *testing.T) {
		requestBody := `{
			"team_name": "Existing Team",
			"members": []
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/teams", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockTeamRepo.EXPECT().
			CreateTeamWithMembers(gomock.Any(), "Existing Team", gomock.Any()).
			Return(uuid.Nil, errors.New("duplicate key error"))

		service.CreateTeam(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("error creating team with members", func(t *testing.T) {
		userID := "user-alice-2"

		requestBody := `{
			"team_name": "Backend Team",
			"members": [
				{"user_id": "` + userID + `", "username": "Alice", "is_active": true}
			]
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/teams", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockTeamRepo.EXPECT().
			CreateTeamWithMembers(gomock.Any(), "Backend Team", gomock.Any()).
			Return(uuid.Nil, errors.New("database error"))

		service.CreateTeam(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
