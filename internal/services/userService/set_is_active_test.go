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
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func TestUserService_SetIsActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockPrReviewersRepo := mocks.NewMockPrReviewersRepositoryInterface(ctrl)
	service := NewUserService(mockUserRepo, mockPrReviewersRepo)

	t.Run("successfully set user active", func(t *testing.T) {
		userID := uuid.NewString()
		user := &domain.User{
			UserId:   userID,
			Username: "Alice",
			IsActive: true,
		}

		requestBody := `{
			"user_id": "` + userID + `",
			"is_active": true
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/users/set-active", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockUserRepo.EXPECT().
			SetUserIsActive(gomock.Any(), userID, true).
			Return(nil)

		mockUserRepo.EXPECT().
			GetUserByID(gomock.Any(), userID).
			Return(user, nil)

		service.SetIsActive(c)

		require.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "user")
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/users/set-active", strings.NewReader("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		service.SetIsActive(c)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var response domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, domain.InvalidRequest, response.Error.Code)
	})

	t.Run("user not found when setting status", func(t *testing.T) {
		userID := uuid.NewString()
		requestBody := `{
			"user_id": "` + userID + `",
			"is_active": true
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/users/set-active", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockUserRepo.EXPECT().
			SetUserIsActive(gomock.Any(), userID, true).
			Return(pgx.ErrNoRows)

		service.SetIsActive(c)

		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("error getting user after update", func(t *testing.T) {
		userID := uuid.NewString()
		requestBody := `{
			"user_id": "` + userID + `",
			"is_active": true
		}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/users/set-active", strings.NewReader(requestBody))
		c.Request.Header.Set("Content-Type", "application/json")

		mockUserRepo.EXPECT().
			SetUserIsActive(gomock.Any(), userID, true).
			Return(nil)

		mockUserRepo.EXPECT().
			GetUserByID(gomock.Any(), userID).
			Return(nil, errors.New("database error"))

		service.SetIsActive(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

