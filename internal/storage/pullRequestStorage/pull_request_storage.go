package pullRequestStorage

import (
	"context"
	"time"

	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/db"
)

type PullRequestStorage struct {
	db db.Querier
}

func NewPullRequestStorage(db db.Querier) *PullRequestStorage {
	return &PullRequestStorage{
		db: db,
	}
}

func (s *PullRequestStorage) GetPullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	var name string
	var authorID string
	var status string
	var createdAt time.Time
	var mergedAt *time.Time

	query := `
		SELECT name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE id = $1`

	err := s.db.QueryRow(ctx, query, prID).Scan(&name, &authorID, &status, &createdAt, &mergedAt)
	if err != nil {
		return nil, err
	}

	pr := &domain.PullRequest{
		PullRequestId:     prID,
		PullRequestName:   name,
		AuthorId:          authorID,
		Status:            domain.PullRequestStatus(status),
		AssignedReviewers: nil,
		CreatedAt:         &createdAt,
		MergedAt:          mergedAt,
	}

	return pr, nil
}

func (s *PullRequestStorage) MergePullRequest(ctx context.Context, prID string) error {
	mergedAt := time.Now()

	query := `
		UPDATE pull_requests
		SET status = $1, merged_at = $2
		WHERE id = $3 AND status != $1`

	_, err := s.db.Exec(ctx, query, string(domain.PullRequestStatusMERGED), mergedAt, prID)
	if err != nil {
		return err
	}

	return nil
}
func (s *PullRequestStorage) SetNeedMoreReviewers(ctx context.Context, prID string, needMore bool) error {
	query := `UPDATE pull_requests SET need_more_reviewers = $1 WHERE id = $2`

	_, err := s.db.Exec(ctx, query, needMore, prID)
	if err != nil {
		return err
	}

	return nil
}

func (s *PullRequestStorage) CreatePullRequestWithReviewers(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string, needMoreReviewers bool) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	prQuery := `INSERT INTO pull_requests (id, name, author_id, status) VALUES ($1, $2, $3, $4)`
	_, err = tx.Exec(ctx, prQuery, pr.PullRequestId, pr.PullRequestName, pr.AuthorId, string(pr.Status))
	if err != nil {
		return err
	}

	reviewerQuery := `
		INSERT INTO pr_reviewers (pull_request_id, reviewer_id, assigned_at)
		VALUES ($1, $2, $3)`

	now := time.Now()
	for _, reviewerID := range reviewerIDs {
		_, err = tx.Exec(ctx, reviewerQuery, pr.PullRequestId, reviewerID, now)
		if err != nil {
			return err
		}
	}

	if needMoreReviewers {
		updateQuery := `UPDATE pull_requests SET need_more_reviewers = $1 WHERE id = $2`
		_, err = tx.Exec(ctx, updateQuery, true, pr.PullRequestId)
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
