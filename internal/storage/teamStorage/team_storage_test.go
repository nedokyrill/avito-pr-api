package teamStorage

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testTeam   = "Backend Team"
	testStrID  = "test-str-id"
)

func TestTeamStorage_GetTeamByName(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully get team with members", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewTeamStorage(mock)
		teamName := testTeam
		teamID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		user1ID := "test-id"
		user2ID := "test-id"

		mock.ExpectQuery("SELECT id FROM teams WHERE name").
			WithArgs(teamName).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(teamID))

		mock.ExpectQuery("SELECT id, name, is_active").
			WithArgs(teamID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "is_active"}).
				AddRow(user1ID, "Alice", true).
				AddRow(user2ID, "Bob", false))

		team, err := storage.GetTeamByName(ctx, teamName)

		require.NoError(t, err)
		require.NotNil(t, team)
		assert.Equal(t, teamName, team.TeamName)
		assert.Len(t, team.Members, 2)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("team not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewTeamStorage(mock)

		mock.ExpectQuery("SELECT id FROM teams WHERE name").
			WithArgs("NonExistent Team").
			WillReturnError(pgx.ErrNoRows)

		team, err := storage.GetTeamByName(ctx, "NonExistent Team")

		assert.Error(t, err)
		assert.Equal(t, ErrTeamNotExists, err)
		assert.Nil(t, team)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTeamStorage_CreateTeamWithMembers(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully create team with members", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewTeamStorage(mock)
		teamName := testTeam
		teamID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		user1ID := testStrID
		user2ID := testStrID

		members := []domain.TeamMember{
			{UserId: user1ID, Username: "Alice", IsActive: true},
			{UserId: user2ID, Username: "Bob", IsActive: false},
		}

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO teams").
			WithArgs(teamName).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(teamID))

		mock.ExpectExec("INSERT INTO users").
			WithArgs(pgxmock.AnyArg(), "Alice", teamID, true).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectExec("INSERT INTO users").
			WithArgs(pgxmock.AnyArg(), "Bob", teamID, false).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectCommit()
		mock.ExpectRollback()

		resultID, err := storage.CreateTeamWithMembers(ctx, teamName, members)

		require.NoError(t, err)
		assert.Equal(t, teamID, resultID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error creating team - rollback", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewTeamStorage(mock)
		teamName := testTeam
		members := []domain.TeamMember{
			{UserId: testStrID, Username: "Alice", IsActive: true},
		}

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO teams").
			WithArgs(teamName).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		resultID, err := storage.CreateTeamWithMembers(ctx, teamName, members)

		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, resultID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error adding user - rollback", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewTeamStorage(mock)
		teamName := testTeam
		teamID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
		user1ID := testStrID

		members := []domain.TeamMember{
			{UserId: user1ID, Username: "Alice", IsActive: true},
		}

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO teams").
			WithArgs(teamName).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(teamID))

		mock.ExpectExec("INSERT INTO users").
			WithArgs(pgxmock.AnyArg(), "Alice", teamID, true).
			WillReturnError(errors.New("user insert error"))

		mock.ExpectRollback()

		resultID, err := storage.CreateTeamWithMembers(ctx, teamName, members)

		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, resultID)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTeamStorage_DeactivateTeamMembers(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := NewTeamStorage(mock)

	t.Run("successfully deactivate specific users with reassignments", func(t *testing.T) {
		teamName := "Backend"
		userID1 := "user-1"
		userID2 := "user-2"
		prID := "pr-123"
		newReviewerID := "user-3"

		mock.ExpectBegin()

		// UPDATE users query
		mock.ExpectQuery(`UPDATE users u`).
			WithArgs(teamName, []string{userID1, userID2}).
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(userID1).AddRow(userID2))

		// DELETE old reviewer
		mock.ExpectExec(`DELETE FROM pr_reviewers`).
			WithArgs(prID, userID1).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		// INSERT new reviewer
		mock.ExpectExec(`INSERT INTO pr_reviewers`).
			WithArgs(prID, newReviewerID).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectCommit()

		reassignments := []domain.ReviewerReassignment{
			{
				PrID:          prID,
				OldReviewerID: userID1,
				NewReviewerID: newReviewerID,
			},
		}

		result, err := storage.DeactivateTeamMembers(context.Background(), teamName, []string{userID1, userID2}, reassignments)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, userID1)
		assert.Contains(t, result, userID2)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
