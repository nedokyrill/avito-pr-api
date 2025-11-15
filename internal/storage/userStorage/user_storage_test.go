package userStorage

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStorage_GetUserByID(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully get user", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewUserStorage(mock)
		userID := "user-123"

		mock.ExpectQuery("SELECT u.name, t.name, u.is_active").
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{"name", "name", "is_active"}).
				AddRow("Alice", "Backend Team", true))

		user, err := storage.GetUserByID(ctx, userID)

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, userID, user.UserId)
		assert.Equal(t, "Alice", user.Username)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewUserStorage(mock)
		userID := "user-456"

		mock.ExpectQuery("SELECT u.name, t.name, u.is_active").
			WithArgs(userID).
			WillReturnError(pgx.ErrNoRows)

		user, err := storage.GetUserByID(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, user)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserStorage_SetUserIsActive(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully set user active", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewUserStorage(mock)
		userID := "user-789"

		mock.ExpectExec("UPDATE users").
			WithArgs(true, userID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = storage.SetUserIsActive(ctx, userID, true)

		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
