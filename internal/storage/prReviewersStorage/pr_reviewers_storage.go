package prReviewersStorage

import (
	"context"
	"time"

	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/db"
)

type PrReviewersStorage struct {
	db db.Querier
}

func NewPrReviewersStorage(db db.Querier) *PrReviewersStorage {
	return &PrReviewersStorage{
		db: db,
	}
}

func (s *PrReviewersStorage) GetAssignedReviewers(ctx context.Context, prID string) ([]string, error) {
	query := `
		SELECT reviewer_id
		FROM pr_reviewers
		WHERE pull_request_id = $1
		ORDER BY assigned_at`

	rows, err := s.db.Query(ctx, query, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewerID string
		if err = rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewerID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reviewers, nil
}

func (s *PrReviewersStorage) GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	query := `
		SELECT pr.id, pr.name, pr.author_id, pr.status
		FROM pull_requests pr
		JOIN pr_reviewers prr ON pr.id = prr.pull_request_id
		WHERE prr.reviewer_id = $1
		ORDER BY pr.created_at DESC`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []domain.PullRequestShort
	for rows.Next() {
		var prID string
		var name string
		var authorID string
		var status string

		if err = rows.Scan(&prID, &name, &authorID, &status); err != nil {
			return nil, err
		}

		prs = append(prs, domain.PullRequestShort{
			PullRequestId:   prID,
			PullRequestName: name,
			AuthorId:        authorID,
			Status:          domain.PullRequestStatus(status),
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}

func (s *PrReviewersStorage) ReassignReviewerAtomic(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	deleteQuery := `DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2`
	_, err = tx.Exec(ctx, deleteQuery, prID, oldReviewerID)
	if err != nil {
		return err
	}

	insertQuery := `
		INSERT INTO pr_reviewers (pull_request_id, reviewer_id, assigned_at)
		VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, insertQuery, prID, newReviewerID, time.Now())
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
