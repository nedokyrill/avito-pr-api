package pullRequestStorage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testID    = "test-id"
	testStrID = "test-str-id"
)

func TestPullRequestStorage_GetPullRequestByID(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully get PR", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPullRequestStorage(mock)
		prID := testID
		authorID := testID
		createdAt := time.Now()
		mergedAt := time.Now().Add(2 * time.Hour)

		mock.ExpectQuery("SELECT name, author_id, status, created_at, merged_at").
			WithArgs(prID).
			WillReturnRows(pgxmock.NewRows([]string{"name", "author_id", "status", "created_at", "merged_at"}).
				AddRow("Add feature", authorID, string(domain.PullRequestStatusMERGED), createdAt, &mergedAt))

		pr, err := storage.GetPullRequestByID(ctx, prID)

		require.NoError(t, err)
		require.NotNil(t, pr)
		assert.Equal(t, prID, pr.PullRequestId)
		assert.NotNil(t, pr.MergedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("PR not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPullRequestStorage(mock)
		prID := testID

		mock.ExpectQuery("SELECT name, author_id, status, created_at, merged_at").
			WithArgs(prID).
			WillReturnError(pgx.ErrNoRows)

		pr, err := storage.GetPullRequestByID(ctx, prID)

		assert.Error(t, err)
		assert.Nil(t, pr)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPullRequestStorage_MergePullRequest(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully merge PR", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPullRequestStorage(mock)
		prID := testID

		mock.ExpectExec("UPDATE pull_requests").
			WithArgs(string(domain.PullRequestStatusMERGED), pgxmock.AnyArg(), prID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = storage.MergePullRequest(ctx, prID)

		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPullRequestStorage_SetNeedMoreReviewers(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully set flag", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPullRequestStorage(mock)
		prID := testID

		mock.ExpectExec("UPDATE pull_requests SET need_more_reviewers").
			WithArgs(true, prID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = storage.SetNeedMoreReviewers(ctx, prID, true)

		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPullRequestStorage_CreatePullRequestWithReviewers(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully create PR with reviewers", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPullRequestStorage(mock)
		prID := testStrID
		authorID := testStrID
		reviewer1ID := testStrID
		reviewer2ID := testStrID

		pr := &domain.PullRequest{
			PullRequestId:   prID,
			PullRequestName: "Add feature",
			AuthorId:        authorID,
			Status:          domain.PullRequestStatusOPEN,
		}

		reviewerIDs := []string{reviewer1ID, reviewer2ID}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO pull_requests").
			WithArgs(pgxmock.AnyArg(), "Add feature", pgxmock.AnyArg(), string(domain.PullRequestStatusOPEN)).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectExec("INSERT INTO pr_reviewers").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectExec("INSERT INTO pr_reviewers").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectCommit()
		mock.ExpectRollback()

		err = storage.CreatePullRequestWithReviewers(ctx, pr, reviewerIDs, false)

		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("successfully create PR with need_more_reviewers flag", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPullRequestStorage(mock)
		prID := testStrID
		authorID := testStrID
		reviewer1ID := testStrID

		pr := &domain.PullRequest{
			PullRequestId:   prID,
			PullRequestName: "Add feature",
			AuthorId:        authorID,
			Status:          domain.PullRequestStatusOPEN,
		}

		reviewerIDs := []string{reviewer1ID}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO pull_requests").
			WithArgs(pgxmock.AnyArg(), "Add feature", pgxmock.AnyArg(), string(domain.PullRequestStatusOPEN)).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectExec("INSERT INTO pr_reviewers").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectExec("UPDATE pull_requests SET need_more_reviewers").
			WithArgs(true, pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		mock.ExpectCommit()
		mock.ExpectRollback()

		err = storage.CreatePullRequestWithReviewers(ctx, pr, reviewerIDs, true)

		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error creating PR - rollback", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPullRequestStorage(mock)
		prID := testStrID
		authorID := testStrID

		pr := &domain.PullRequest{
			PullRequestId:   prID,
			PullRequestName: "Add feature",
			AuthorId:        authorID,
			Status:          domain.PullRequestStatusOPEN,
		}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO pull_requests").
			WithArgs(pgxmock.AnyArg(), "Add feature", pgxmock.AnyArg(), string(domain.PullRequestStatusOPEN)).
			WillReturnError(errors.New("database error"))

		mock.ExpectRollback()

		err = storage.CreatePullRequestWithReviewers(ctx, pr, []string{}, false)

		assert.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error adding reviewer - rollback", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPullRequestStorage(mock)
		prID := testStrID
		authorID := testStrID
		reviewer1ID := testStrID

		pr := &domain.PullRequest{
			PullRequestId:   prID,
			PullRequestName: "Add feature",
			AuthorId:        authorID,
			Status:          domain.PullRequestStatusOPEN,
		}

		reviewerIDs := []string{reviewer1ID}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO pull_requests").
			WithArgs(pgxmock.AnyArg(), "Add feature", pgxmock.AnyArg(), string(domain.PullRequestStatusOPEN)).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectExec("INSERT INTO pr_reviewers").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errors.New("reviewer insert error"))

		mock.ExpectRollback()

		err = storage.CreatePullRequestWithReviewers(ctx, pr, reviewerIDs, false)

		assert.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
