package prReviewersStorage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
)

var ErrInvalidUUID = errors.New(domain.InvalidUUIDErr)

type PrReviewersStorage struct {
	db *pgxpool.Pool
}

func NewPrReviewersStorage(db *pgxpool.Pool) *PrReviewersStorage {
	return &PrReviewersStorage{
		db: db,
	}
}

func (s *PrReviewersStorage) RemoveReviewer(ctx context.Context, prID string, reviewerID string) error {
	prUUID, err := uuid.Parse(prID)
	if err != nil {
		return ErrInvalidUUID
	}

	reviewerUUID, err := uuid.Parse(reviewerID)
	if err != nil {
		return ErrInvalidUUID
	}

	query := `DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2`

	_, err = s.db.Exec(ctx, query, prUUID, reviewerUUID)
	if err != nil {
		return err
	}

	return nil
}

func (s *PrReviewersStorage) AddReviewer(ctx context.Context, prID string, reviewerID string) error {
	prUUID, err := uuid.Parse(prID)
	if err != nil {
		return ErrInvalidUUID
	}

	reviewerUUID, err := uuid.Parse(reviewerID)
	if err != nil {
		return ErrInvalidUUID
	}

	query := `
		INSERT INTO pr_reviewers (pull_request_id, reviewer_id, assigned_at)
		VALUES ($1, $2, $3)`

	_, err = s.db.Exec(ctx, query, prUUID, reviewerUUID, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (s *PrReviewersStorage) GetAssignedReviewers(ctx context.Context, prID string) ([]string, error) {
	prUUID, err := uuid.Parse(prID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	query := `
		SELECT reviewer_id
		FROM pr_reviewers
		WHERE pull_request_id = $1
		ORDER BY assigned_at`

	rows, err := s.db.Query(ctx, query, prUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewerUUID uuid.UUID
		if err = rows.Scan(&reviewerUUID); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewerUUID.String())
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reviewers, nil
}

func (s *PrReviewersStorage) GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	query := `
		SELECT pr.id, pr.name, pr.author_id, pr.status
		FROM pull_requests pr
		JOIN pr_reviewers prr ON pr.id = prr.pull_request_id
		WHERE prr.reviewer_id = $1
		ORDER BY pr.created_at DESC`

	rows, err := s.db.Query(ctx, query, userUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []domain.PullRequestShort
	for rows.Next() {
		var prUUID uuid.UUID
		var name string
		var authorUUID uuid.UUID
		var status string

		if err = rows.Scan(&prUUID, &name, &authorUUID, &status); err != nil {
			return nil, err
		}

		prs = append(prs, domain.PullRequestShort{
			PullRequestId:   prUUID.String(),
			PullRequestName: name,
			AuthorId:        authorUUID.String(),
			Status:          domain.PullRequestStatus(status),
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}
