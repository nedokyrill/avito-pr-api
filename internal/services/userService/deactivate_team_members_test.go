package userService

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testTeamNameBackend = "Backend"
	testUserID1         = "user-1"
	testUserID2         = "user-2"
	testUserID3         = "user-3"
	testPRID123         = "pr-123"
)

func TestUserService_DeactivateTeamMembers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockPrReviewersRepo := mocks.NewMockPrReviewersRepositoryInterface(ctrl)
	mockTeamRepo := mocks.NewMockTeamRepositoryInterface(ctrl)
	service := NewUserService(mockUserRepo, mockPrReviewersRepo, mockTeamRepo)

	t.Run("successfully deactivate team members with reassignments", func(t *testing.T) {
		teamName := testTeamNameBackend
		userID1 := testUserID1
		userID2 := testUserID2
		prID := testPRID123
		authorID := testUserID3
		userID4 := "user-4" // Дополнительный активный участник для замены ревьювера

		requestBody := `{
			"team_name": "` + teamName + `",
			"user_ids": ["` + userID1 + `", "` + userID2 + `"]
		}`

		team := &domain.Team{
			TeamName: teamName,
			Members: []domain.TeamMember{
				{UserId: authorID, Username: "Alice", IsActive: true},
				{UserId: userID1, Username: "Bob", IsActive: true},
				{UserId: userID2, Username: "Charlie", IsActive: true},
				{UserId: userID4, Username: "David", IsActive: true}, // Активный участник для замены
			},
		}

		prs := []domain.PullRequestShort{
			{
				PullRequestId:   prID,
				PullRequestName: "Feature PR",
				AuthorId:        authorID,
				Status:          domain.PullRequestStatusOPEN,
			},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/users/deactivate", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockTeamRepo.EXPECT().
			GetTeamByName(gomock.Any(), teamName).
			Return(team, nil)

		mockPrReviewersRepo.EXPECT().
			GetPRsByReviewer(gomock.Any(), userID1).
			Return(prs, nil)

		mockPrReviewersRepo.EXPECT().
			GetPRsByReviewer(gomock.Any(), userID2).
			Return([]domain.PullRequestShort{}, nil)

		mockPrReviewersRepo.EXPECT().
			GetAssignedReviewers(gomock.Any(), prID).
			Return([]string{userID1}, nil)

		mockTeamRepo.EXPECT().
			DeactivateTeamMembers(gomock.Any(), teamName, []string{userID1, userID2}, gomock.Any()).
			Return([]string{userID1, userID2}, nil)

		service.DeactivateTeamMembers(c)

		require.Equal(t, http.StatusOK, w.Code)
		var response domain.DeactivateTeamMembersResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.DeactivatedUserIDs, 2)
		assert.Contains(t, response.DeactivatedUserIDs, userID1)
		assert.Contains(t, response.DeactivatedUserIDs, userID2)
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/users/deactivate", strings.NewReader("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		service.DeactivateTeamMembers(c)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var response domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, domain.InvalidRequest, response.Error.Code)
	})

	t.Run("team not found", func(t *testing.T) {
		teamName := "NonExistent"

		requestBody := `{
			"team_name": "` + teamName + `",
			"user_ids": ["` + testUserID1 + `"]
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/users/deactivate", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockTeamRepo.EXPECT().
			GetTeamByName(gomock.Any(), teamName).
			Return(nil, pgx.ErrNoRows)

		service.DeactivateTeamMembers(c)

		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("cannot deactivate all team members", func(t *testing.T) {
		teamName := testTeamNameBackend
		userID1 := testUserID1
		userID2 := testUserID2

		requestBody := `{
			"team_name": "` + teamName + `",
			"user_ids": ["` + userID1 + `", "` + userID2 + `"]
		}`

		team := &domain.Team{
			TeamName: teamName,
			Members: []domain.TeamMember{
				{UserId: userID1, Username: "Bob", IsActive: true},
				{UserId: userID2, Username: "Charlie", IsActive: true},
			},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/users/deactivate", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockTeamRepo.EXPECT().
			GetTeamByName(gomock.Any(), teamName).
			Return(team, nil)

		service.DeactivateTeamMembers(c)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var response domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, domain.InvalidRequest, response.Error.Code)
	})

	t.Run("user not member of team", func(t *testing.T) {
		teamName := testTeamNameBackend
		userID1 := testUserID1
		userID2 := "user-from-other-team"

		requestBody := `{
			"team_name": "` + teamName + `",
			"user_ids": ["` + userID1 + `", "` + userID2 + `"]
		}`

		team := &domain.Team{
			TeamName: teamName,
			Members: []domain.TeamMember{
				{UserId: userID1, Username: "Bob", IsActive: true},
				{UserId: "user-3", Username: "Alice", IsActive: true},
			},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/users/deactivate", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockTeamRepo.EXPECT().
			GetTeamByName(gomock.Any(), teamName).
			Return(team, nil)

		service.DeactivateTeamMembers(c)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var response domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, domain.InvalidRequest, response.Error.Code)
		assert.Contains(t, response.Error.Message, "cannot deactivate all team members")
	})

	t.Run("error getting team", func(t *testing.T) {
		teamName := testTeamNameBackend

		requestBody := `{
			"team_name": "` + teamName + `",
			"user_ids": ["` + testUserID1 + `"]
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/users/deactivate", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockTeamRepo.EXPECT().
			GetTeamByName(gomock.Any(), teamName).
			Return(nil, errors.New("database error"))

		service.DeactivateTeamMembers(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("error deactivating users", func(t *testing.T) {
		teamName := testTeamNameBackend
		userID1 := testUserID1

		requestBody := `{
			"team_name": "` + teamName + `",
			"user_ids": ["` + userID1 + `"]
		}`

		team := &domain.Team{
			TeamName: teamName,
			Members: []domain.TeamMember{
				{UserId: userID1, Username: "Bob", IsActive: true},
				{UserId: "user-2", Username: "Alice", IsActive: true},
			},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/users/deactivate", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockTeamRepo.EXPECT().
			GetTeamByName(gomock.Any(), teamName).
			Return(team, nil)

		mockPrReviewersRepo.EXPECT().
			GetPRsByReviewer(gomock.Any(), userID1).
			Return([]domain.PullRequestShort{}, nil)

		mockTeamRepo.EXPECT().
			DeactivateTeamMembers(gomock.Any(), teamName, gomock.Any(), gomock.Any()).
			Return(nil, errors.New("database error"))

		service.DeactivateTeamMembers(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

