package pullRequestService

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/internal/storage/mocks"
	"github.com/stretchr/testify/require"
)

func TestPullRequestService_ReassignReviewer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPrRepo := mocks.NewMockPullRequestRepositoryInterface(ctrl)
	mockPrReviewersRepo := mocks.NewMockPrReviewersRepositoryInterface(ctrl)
	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockTeamRepo := mocks.NewMockTeamRepositoryInterface(ctrl)

	service := NewPullRequestService(mockPrRepo, mockPrReviewersRepo, mockUserRepo, mockTeamRepo)

	t.Run("successfully reassign reviewer", func(t *testing.T) {
		prID := uuid.NewString()
		oldReviewerID := uuid.NewString()
		newReviewerID := uuid.NewString()
		authorID := uuid.NewString()

		pr := &domain.PullRequest{
			PullRequestId:   prID,
			PullRequestName: "Add feature",
			AuthorId:        authorID,
			Status:          domain.PullRequestStatusOPEN,
		}

		oldReviewer := &domain.User{
			UserId:   oldReviewerID,
			Username: "Bob",
			TeamName: "Backend",
			IsActive: true,
		}

		team := &domain.Team{
			TeamName: "Backend",
			Members: []domain.TeamMember{
				{UserId: authorID, Username: "Alice", IsActive: true},
				{UserId: oldReviewerID, Username: "Bob", IsActive: true},
				{UserId: newReviewerID, Username: "Charlie", IsActive: true},
			},
		}

		requestBody := `{
			"pull_request_id": "` + prID + `",
			"old_user_id": "` + oldReviewerID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests/reassign", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockPrRepo.EXPECT().GetPullRequestByID(gomock.Any(), prID).Return(pr, nil)
		mockPrReviewersRepo.EXPECT().GetAssignedReviewers(gomock.Any(), prID).Return([]string{oldReviewerID}, nil)
		mockUserRepo.EXPECT().GetUserByID(gomock.Any(), oldReviewerID).Return(oldReviewer, nil)
		mockTeamRepo.EXPECT().GetTeamByName(gomock.Any(), "Backend").Return(team, nil)
		mockPrReviewersRepo.EXPECT().ReassignReviewerAtomic(gomock.Any(), prID, oldReviewerID, gomock.Any()).Return(nil)
		mockPrReviewersRepo.EXPECT().GetAssignedReviewers(gomock.Any(), prID).Return([]string{newReviewerID}, nil)

		service.ReassignReviewer(c)

		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests/reassign", strings.NewReader("invalid"))
		c.Request.Header.Set("Content-Type", "application/json")

		service.ReassignReviewer(c)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PR not found", func(t *testing.T) {
		prID := uuid.NewString()
		reviewerID := uuid.NewString()

		requestBody := `{
			"pull_request_id": "` + prID + `",
			"old_user_id": "` + reviewerID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests/reassign", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockPrRepo.EXPECT().GetPullRequestByID(gomock.Any(), prID).Return(nil, pgx.ErrNoRows)

		service.ReassignReviewer(c)

		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("PR already merged", func(t *testing.T) {
		prID := uuid.NewString()
		reviewerID := uuid.NewString()

		pr := &domain.PullRequest{
			PullRequestId:   prID,
			PullRequestName: "Add feature",
			Status:          domain.PullRequestStatusMERGED,
		}

		requestBody := `{
			"pull_request_id": "` + prID + `",
			"old_user_id": "` + reviewerID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests/reassign", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockPrRepo.EXPECT().GetPullRequestByID(gomock.Any(), prID).Return(pr, nil)

		service.ReassignReviewer(c)

		require.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("reviewer not assigned", func(t *testing.T) {
		prID := uuid.NewString()
		reviewerID := uuid.NewString()
		otherReviewerID := uuid.NewString()

		pr := &domain.PullRequest{
			PullRequestId:   prID,
			PullRequestName: "Add feature",
			Status:          domain.PullRequestStatusOPEN,
		}

		requestBody := `{
			"pull_request_id": "` + prID + `",
			"old_user_id": "` + reviewerID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests/reassign", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockPrRepo.EXPECT().GetPullRequestByID(gomock.Any(), prID).Return(pr, nil)
		mockPrReviewersRepo.EXPECT().GetAssignedReviewers(gomock.Any(), prID).Return([]string{otherReviewerID}, nil)

		service.ReassignReviewer(c)

		require.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("error getting user", func(t *testing.T) {
		prID := uuid.NewString()
		reviewerID := uuid.NewString()

		pr := &domain.PullRequest{
			PullRequestId:   prID,
			PullRequestName: "Add feature",
			Status:          domain.PullRequestStatusOPEN,
		}

		requestBody := `{
			"pull_request_id": "` + prID + `",
			"old_user_id": "` + reviewerID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests/reassign", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockPrRepo.EXPECT().GetPullRequestByID(gomock.Any(), prID).Return(pr, nil)
		mockPrReviewersRepo.EXPECT().GetAssignedReviewers(gomock.Any(), prID).Return([]string{reviewerID}, nil)
		mockUserRepo.EXPECT().GetUserByID(gomock.Any(), reviewerID).Return(nil, errors.New("db error"))

		service.ReassignReviewer(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("error reassigning reviewer", func(t *testing.T) {
		prID := uuid.NewString()
		oldReviewerID := uuid.NewString()
		newReviewerID := uuid.NewString()
		authorID := uuid.NewString()

		pr := &domain.PullRequest{
			PullRequestId:   prID,
			PullRequestName: "Add feature",
			AuthorId:        authorID,
			Status:          domain.PullRequestStatusOPEN,
		}

		oldReviewer := &domain.User{
			UserId:   oldReviewerID,
			Username: "Bob",
			TeamName: "Backend",
			IsActive: true,
		}

		team := &domain.Team{
			TeamName: "Backend",
			Members: []domain.TeamMember{
				{UserId: authorID, Username: "Alice", IsActive: true},
				{UserId: oldReviewerID, Username: "Bob", IsActive: true},
				{UserId: newReviewerID, Username: "Charlie", IsActive: true},
			},
		}

		requestBody := `{
			"pull_request_id": "` + prID + `",
			"old_user_id": "` + oldReviewerID + `"
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/pull-requests/reassign", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockPrRepo.EXPECT().GetPullRequestByID(gomock.Any(), prID).Return(pr, nil)
		mockPrReviewersRepo.EXPECT().GetAssignedReviewers(gomock.Any(), prID).Return([]string{oldReviewerID}, nil)
		mockUserRepo.EXPECT().GetUserByID(gomock.Any(), oldReviewerID).Return(oldReviewer, nil)
		mockTeamRepo.EXPECT().GetTeamByName(gomock.Any(), "Backend").Return(team, nil)
		mockPrReviewersRepo.EXPECT().ReassignReviewerAtomic(gomock.Any(), prID, oldReviewerID, gomock.Any()).Return(errors.New("db error"))

		service.ReassignReviewer(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
