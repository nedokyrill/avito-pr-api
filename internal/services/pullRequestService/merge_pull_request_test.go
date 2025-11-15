package pullRequestService

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullRequestService_MergePullRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPrRepo := mocks.NewMockPullRequestRepositoryInterface(ctrl)
	mockPrReviewersRepo := mocks.NewMockPrReviewersRepositoryInterface(ctrl)
	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockTeamRepo := mocks.NewMockTeamRepositoryInterface(ctrl)

	service := NewPullRequestService(mockPrRepo, mockPrReviewersRepo, mockUserRepo, mockTeamRepo)

	t.Run("successfully merge PR", func(t *testing.T) {
		prID := testStrID
		authorID := testStrID

		pr := &domain.PullRequest{
			PullRequestId:   prID,
			PullRequestName: "Add feature",
			AuthorId:        authorID,
			Status:          domain.PullRequestStatusOPEN,
		}

		requestBody := `{
			"pull_request_id": "` + prID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests/merge", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockPrRepo.EXPECT().GetPullRequestByID(gomock.Any(), prID).Return(pr, nil)
		mockPrRepo.EXPECT().MergePullRequest(gomock.Any(), prID).Return(nil)
		mockPrReviewersRepo.EXPECT().GetAssignedReviewers(gomock.Any(), prID).Return([]string{}, nil)

		service.MergePullRequest(c)

		require.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "pr")
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests/merge", strings.NewReader("invalid"))
		c.Request.Header.Set("Content-Type", "application/json")

		service.MergePullRequest(c)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PR not found", func(t *testing.T) {
		prID := testStrID

		requestBody := `{
			"pull_request_id": "` + prID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests/merge", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockPrRepo.EXPECT().GetPullRequestByID(gomock.Any(), prID).Return(nil, pgx.ErrNoRows)

		service.MergePullRequest(c)

		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("error getting reviewers", func(t *testing.T) {
		prID := testStrID
		createdAt := time.Now().Add(-2 * time.Hour)

		pr := &domain.PullRequest{
			PullRequestId:   prID,
			PullRequestName: "Test PR",
			AuthorId:        "author1",
			Status:          domain.PullRequestStatusOPEN,
			CreatedAt:       &createdAt,
		}

		requestBody := `{
			"pull_request_id": "` + prID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests/merge", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockPrRepo.EXPECT().GetPullRequestByID(gomock.Any(), prID).Return(pr, nil)
		mockPrRepo.EXPECT().MergePullRequest(gomock.Any(), prID).Return(nil)
		mockPrReviewersRepo.EXPECT().GetAssignedReviewers(gomock.Any(), prID).Return(nil, errors.New("db error"))

		service.MergePullRequest(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("error merging PR", func(t *testing.T) {
		prID := testStrID
		createdAt := time.Now().Add(-2 * time.Hour)

		pr := &domain.PullRequest{
			PullRequestId:   prID,
			PullRequestName: "Test PR",
			AuthorId:        "author1",
			Status:          domain.PullRequestStatusOPEN,
			CreatedAt:       &createdAt,
		}

		requestBody := `{
			"pull_request_id": "` + prID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests/merge", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockPrRepo.EXPECT().GetPullRequestByID(gomock.Any(), prID).Return(pr, nil)
		mockPrRepo.EXPECT().MergePullRequest(gomock.Any(), prID).Return(errors.New("db error"))

		service.MergePullRequest(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("PR already merged", func(t *testing.T) {
		prID := testStrID
		authorID := testStrID

		pr := &domain.PullRequest{
			PullRequestId:   prID,
			PullRequestName: "Add feature",
			AuthorId:        authorID,
			Status:          domain.PullRequestStatusMERGED,
		}

		requestBody := `{
			"pull_request_id": "` + prID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests/merge", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockPrRepo.EXPECT().GetPullRequestByID(gomock.Any(), prID).Return(pr, nil)
		mockPrReviewersRepo.EXPECT().GetAssignedReviewers(gomock.Any(), prID).Return([]string{}, nil)

		service.MergePullRequest(c)

		require.Equal(t, http.StatusOK, w.Code)
	})
}

