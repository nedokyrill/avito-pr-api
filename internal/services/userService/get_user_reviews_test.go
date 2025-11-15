package userService

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserService_GetUserReviews(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockPrReviewersRepo := mocks.NewMockPrReviewersRepositoryInterface(ctrl)
	service := NewUserService(mockUserRepo, mockPrReviewersRepo, nil)

	t.Run("successfully get user reviews", func(t *testing.T) {
		userID := testUserIDStr
		user := &domain.User{
			UserId:   userID,
			Username: "Alice",
			IsActive: true,
		}

		prs := []domain.PullRequestShort{
			{
				PullRequestId:   testUserIDStr,
				PullRequestName: "Feature A",
				AuthorId:        testUserIDStr,
				Status:          domain.PullRequestStatusOPEN,
			},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users/reviews?user_id="+userID, nil)

		mockUserRepo.EXPECT().
			GetUserByID(gomock.Any(), userID).
			Return(user, nil)

		mockPrReviewersRepo.EXPECT().
			GetPRsByReviewer(gomock.Any(), userID).
			Return(prs, nil)

		service.GetUserReviews(c)

		require.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, userID, response["user_id"])
		assert.Contains(t, response, "pull_requests")
	})

	t.Run("missing user_id parameter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users/reviews", nil)

		service.GetUserReviews(c)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var response domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, domain.InvalidRequest, response.Error.Code)
	})

	t.Run("user not found", func(t *testing.T) {
		userID := testUserIDStr

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users/reviews?user_id="+userID, nil)

		mockUserRepo.EXPECT().
			GetUserByID(gomock.Any(), userID).
			Return(nil, pgx.ErrNoRows)

		service.GetUserReviews(c)

		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("error getting reviews", func(t *testing.T) {
		userID := testUserIDStr
		user := &domain.User{
			UserId:   userID,
			Username: "Alice",
			IsActive: true,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users/reviews?user_id="+userID, nil)

		mockUserRepo.EXPECT().
			GetUserByID(gomock.Any(), userID).
			Return(user, nil)

		mockPrReviewersRepo.EXPECT().
			GetPRsByReviewer(gomock.Any(), userID).
			Return(nil, errors.New("database error"))

		service.GetUserReviews(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

