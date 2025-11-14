package userStorage

import (
	"context"
	"testing"

	"github.com/google/uuid"
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
		userID := uuid.New()

		mock.ExpectQuery("SELECT u.name, t.name, u.is_active").
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{"name", "name", "is_active"}).
				AddRow("Alice", "Backend Team", true))

		user, err := storage.GetUserByID(ctx, userID.String())

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, userID.String(), user.UserId)
		assert.Equal(t, "Alice", user.Username)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid UUID", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewUserStorage(mock)

		user, err := storage.GetUserByID(ctx, "invalid-uuid")

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidUUID, err)
		assert.Nil(t, user)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewUserStorage(mock)
		userID := uuid.New()

		mock.ExpectQuery("SELECT u.name, t.name, u.is_active").
			WithArgs(userID).
			WillReturnError(pgx.ErrNoRows)

		user, err := storage.GetUserByID(ctx, userID.String())

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
		userID := uuid.New()

		mock.ExpectExec("UPDATE users").
			WithArgs(true, userID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = storage.SetUserIsActive(ctx, userID.String(), true)

		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid UUID", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewUserStorage(mock)

		err = storage.SetUserIsActive(ctx, "invalid-uuid", true)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidUUID, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
