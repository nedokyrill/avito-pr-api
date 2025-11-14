package pullRequestService

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func TestPullRequestService_CreatePullRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPrRepo := mocks.NewMockPullRequestRepositoryInterface(ctrl)
	mockPrReviewersRepo := mocks.NewMockPrReviewersRepositoryInterface(ctrl)
	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockTeamRepo := mocks.NewMockTeamRepositoryInterface(ctrl)

	service := NewPullRequestService(mockPrRepo, mockPrReviewersRepo, mockUserRepo, mockTeamRepo)

	t.Run("successfully create PR with reviewers", func(t *testing.T) {
		prID := uuid.NewString()
		authorID := uuid.NewString()
		reviewer1ID := uuid.NewString()
		reviewer2ID := uuid.NewString()

		author := &domain.User{
			UserId:   authorID,
			Username: "Alice",
			TeamName: "Backend",
			IsActive: true,
		}

		team := &domain.Team{
			TeamName: "Backend",
			Members: []domain.TeamMember{
				{UserId: authorID, Username: "Alice", IsActive: true},
				{UserId: reviewer1ID, Username: "Bob", IsActive: true},
				{UserId: reviewer2ID, Username: "Charlie", IsActive: true},
			},
		}

		requestBody := `{
			"pull_request_id": "` + prID + `",
			"pull_request_name": "Add feature",
			"author_id": "` + authorID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockUserRepo.EXPECT().GetUserByID(gomock.Any(), authorID).Return(author, nil)
		mockTeamRepo.EXPECT().GetTeamByName(gomock.Any(), "Backend").Return(team, nil)
		mockPrRepo.EXPECT().CreatePullRequestWithReviewers(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		service.CreatePullRequest(c)

		require.Equal(t, http.StatusCreated, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "pr")
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests", strings.NewReader("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		service.CreatePullRequest(c)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("author not found", func(t *testing.T) {
		prID := uuid.NewString()
		authorID := uuid.NewString()

		requestBody := `{
			"pull_request_id": "` + prID + `",
			"pull_request_name": "Add feature",
			"author_id": "` + authorID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockUserRepo.EXPECT().GetUserByID(gomock.Any(), authorID).Return(nil, pgx.ErrNoRows)

		service.CreatePullRequest(c)

		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("error getting team", func(t *testing.T) {
		prID := uuid.NewString()
		authorID := uuid.NewString()

		author := &domain.User{
			UserId:   authorID,
			Username: "Alice",
			TeamName: "Backend",
			IsActive: true,
		}

		requestBody := `{
			"pull_request_id": "` + prID + `",
			"pull_request_name": "Add feature",
			"author_id": "` + authorID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockUserRepo.EXPECT().GetUserByID(gomock.Any(), authorID).Return(author, nil)
		mockTeamRepo.EXPECT().GetTeamByName(gomock.Any(), "Backend").Return(nil, errors.New("db error"))

		service.CreatePullRequest(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("PR already exists", func(t *testing.T) {
		prID := uuid.NewString()
		authorID := uuid.NewString()
		reviewer1ID := uuid.NewString()

		author := &domain.User{
			UserId:   authorID,
			Username: "Alice",
			TeamName: "Backend",
			IsActive: true,
		}

		team := &domain.Team{
			TeamName: "Backend",
			Members: []domain.TeamMember{
				{UserId: authorID, Username: "Alice", IsActive: true},
				{UserId: reviewer1ID, Username: "Bob", IsActive: true},
			},
		}

		requestBody := `{
			"pull_request_id": "` + prID + `",
			"pull_request_name": "Add feature",
			"author_id": "` + authorID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockUserRepo.EXPECT().GetUserByID(gomock.Any(), authorID).Return(author, nil)
		mockTeamRepo.EXPECT().GetTeamByName(gomock.Any(), "Backend").Return(team, nil)

		pgErr := &pgconn.PgError{Code: "23505"}
		mockPrRepo.EXPECT().CreatePullRequestWithReviewers(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(pgErr)

		service.CreatePullRequest(c)

		require.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("error creating PR with reviewers", func(t *testing.T) {
		prID := uuid.NewString()
		authorID := uuid.NewString()
		reviewer1ID := uuid.NewString()

		author := &domain.User{
			UserId:   authorID,
			Username: "Alice",
			TeamName: "Backend",
			IsActive: true,
		}

		team := &domain.Team{
			TeamName: "Backend",
			Members: []domain.TeamMember{
				{UserId: authorID, Username: "Alice", IsActive: true},
				{UserId: reviewer1ID, Username: "Bob", IsActive: true},
			},
		}

		requestBody := `{
			"pull_request_id": "` + prID + `",
			"pull_request_name": "Add feature",
			"author_id": "` + authorID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockUserRepo.EXPECT().GetUserByID(gomock.Any(), authorID).Return(author, nil)
		mockTeamRepo.EXPECT().GetTeamByName(gomock.Any(), "Backend").Return(team, nil)
		mockPrRepo.EXPECT().CreatePullRequestWithReviewers(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("db error"))

		service.CreatePullRequest(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
