package prReviewersStorage

import (
	"context"
	"errors"
	"testing"

	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testID     = "test-id"
	testStrID  = "test-str-id"
)

func TestPrReviewersStorage_GetAssignedReviewers(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully get reviewers", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPrReviewersStorage(mock)
		prID := testID
		reviewer1ID := testID
		reviewer2ID := testID

		mock.ExpectQuery("SELECT reviewer_id").
			WithArgs(prID).
			WillReturnRows(pgxmock.NewRows([]string{"reviewer_id"}).
				AddRow(reviewer1ID).
				AddRow(reviewer2ID))

		reviewers, err := storage.GetAssignedReviewers(ctx, prID)

		require.NoError(t, err)
		require.Len(t, reviewers, 2)
		assert.Equal(t, reviewer1ID, reviewers[0])
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no reviewers", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPrReviewersStorage(mock)
		prID := testID

		mock.ExpectQuery("SELECT reviewer_id").
			WithArgs(prID).
			WillReturnRows(pgxmock.NewRows([]string{"reviewer_id"}))

		reviewers, err := storage.GetAssignedReviewers(ctx, prID)

		require.NoError(t, err)
		assert.Empty(t, reviewers)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPrReviewersStorage_GetPRsByReviewer(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully get PRs", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPrReviewersStorage(mock)
		reviewerID := testID
		pr1ID := testID
		author1ID := testID

		mock.ExpectQuery("SELECT pr.id, pr.name, pr.author_id, pr.status").
			WithArgs(reviewerID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "author_id", "status"}).
				AddRow(pr1ID, "Feature A", author1ID, string(domain.PullRequestStatusOPEN)))

		prs, err := storage.GetPRsByReviewer(ctx, reviewerID)

		require.NoError(t, err)
		require.Len(t, prs, 1)
		assert.Equal(t, pr1ID, prs[0].PullRequestId)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no PRs", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPrReviewersStorage(mock)
		reviewerID := testID

		mock.ExpectQuery("SELECT pr.id, pr.name, pr.author_id, pr.status").
			WithArgs(reviewerID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "author_id", "status"}))

		prs, err := storage.GetPRsByReviewer(ctx, reviewerID)

		require.NoError(t, err)
		assert.Empty(t, prs)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPrReviewersStorage_ReassignReviewerAtomic(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully reassign reviewer", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPrReviewersStorage(mock)
		prID := testStrID
		oldReviewerID := testStrID
		newReviewerID := testStrID

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM pr_reviewers").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		mock.ExpectExec("INSERT INTO pr_reviewers").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectCommit()

		err = storage.ReassignReviewerAtomic(ctx, prID, oldReviewerID, newReviewerID)

		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error removing old reviewer - rollback", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPrReviewersStorage(mock)
		prID := testStrID
		oldReviewerID := testStrID
		newReviewerID := testStrID

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM pr_reviewers").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errors.New("delete error"))

		mock.ExpectRollback()

		err = storage.ReassignReviewerAtomic(ctx, prID, oldReviewerID, newReviewerID)

		assert.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error adding new reviewer - rollback", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPrReviewersStorage(mock)
		prID := testStrID
		oldReviewerID := testStrID
		newReviewerID := testStrID

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM pr_reviewers").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		mock.ExpectExec("INSERT INTO pr_reviewers").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errors.New("insert error"))

		mock.ExpectRollback()

		err = storage.ReassignReviewerAtomic(ctx, prID, oldReviewerID, newReviewerID)

		assert.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error on commit", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		storage := NewPrReviewersStorage(mock)
		prID := testStrID
		oldReviewerID := testStrID
		newReviewerID := testStrID

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM pr_reviewers").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		mock.ExpectExec("INSERT INTO pr_reviewers").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectCommit().WillReturnError(errors.New("commit error"))
		mock.ExpectRollback()

		err = storage.ReassignReviewerAtomic(ctx, prID, oldReviewerID, newReviewerID)

		assert.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
