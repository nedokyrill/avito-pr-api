package pullRequestStorage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/db"
)

var ErrInvalidUUID = errors.New(domain.InvalidUUIDErr)

type PullRequestStorage struct {
	db db.Querier
}

func NewPullRequestStorage(db db.Querier) *PullRequestStorage {
	return &PullRequestStorage{
		db: db,
	}
}

func (s *PullRequestStorage) GetPullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	prUUID, err := uuid.Parse(prID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	var name string
	var authorUUID uuid.UUID
	var status string
	var createdAt time.Time
	var mergedAt *time.Time

	query := `
		SELECT name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE id = $1`

	err = s.db.QueryRow(ctx, query, prUUID).Scan(&name, &authorUUID, &status, &createdAt, &mergedAt)
	if err != nil {
		return nil, err
	}

	pr := &domain.PullRequest{
		PullRequestId:     prID,
		PullRequestName:   name,
		AuthorId:          authorUUID.String(),
		Status:            domain.PullRequestStatus(status),
		AssignedReviewers: nil,
		CreatedAt:         &createdAt,
		MergedAt:          mergedAt,
	}

	return pr, nil
}

func (s *PullRequestStorage) MergePullRequest(ctx context.Context, prID string) error {
	prUUID, err := uuid.Parse(prID)
	if err != nil {
		return ErrInvalidUUID
	}

	mergedAt := time.Now()

	query := `
		UPDATE pull_requests
		SET status = $1, merged_at = $2
		WHERE id = $3 AND status != $1`

	_, err = s.db.Exec(ctx, query, string(domain.PullRequestStatusMERGED), mergedAt, prUUID)
	if err != nil {
		return err
	}

	return nil
}
func (s *PullRequestStorage) SetNeedMoreReviewers(ctx context.Context, prID string, needMore bool) error {
	prUUID, err := uuid.Parse(prID)
	if err != nil {
		return ErrInvalidUUID
	}

	query := `UPDATE pull_requests SET need_more_reviewers = $1 WHERE id = $2`

	_, err = s.db.Exec(ctx, query, needMore, prUUID)
	if err != nil {
		return err
	}

	return nil
}

func (s *PullRequestStorage) CreatePullRequestWithReviewers(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string, needMoreReviewers bool) error {
	prUUID, err := uuid.Parse(pr.PullRequestId)
	if err != nil {
		return ErrInvalidUUID
	}

	authorUUID, err := uuid.Parse(pr.AuthorId)
	if err != nil {
		return ErrInvalidUUID
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	prQuery := `INSERT INTO pull_requests (id, name, author_id, status) VALUES ($1, $2, $3, $4)`
	_, err = tx.Exec(ctx, prQuery, prUUID, pr.PullRequestName, authorUUID, string(pr.Status))
	if err != nil {
		return err
	}

	reviewerQuery := `
		INSERT INTO pr_reviewers (pull_request_id, reviewer_id, assigned_at)
		VALUES ($1, $2, $3)`

	now := time.Now()
	for _, reviewerID := range reviewerIDs {
		reviewerUUID, parseErr := uuid.Parse(reviewerID)
		if parseErr != nil {
			return parseErr
		}

		_, err = tx.Exec(ctx, reviewerQuery, prUUID, reviewerUUID, now)
		if err != nil {
			return err
		}
	}

	if needMoreReviewers {
		updateQuery := `UPDATE pull_requests SET need_more_reviewers = $1 WHERE id = $2`
		_, err = tx.Exec(ctx, updateQuery, true, prUUID)
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
