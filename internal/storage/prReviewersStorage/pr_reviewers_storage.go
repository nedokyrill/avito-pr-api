package prReviewersStorage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/db"
)

var ErrInvalidUUID = errors.New(domain.InvalidUUIDErr)

type PrReviewersStorage struct {
	db db.Querier
}

func NewPrReviewersStorage(db db.Querier) *PrReviewersStorage {
	return &PrReviewersStorage{
		db: db,
	}
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

// ReassignReviewerAtomic атомарно переназначает ревьюера (удаляет старого и добавляет нового) в транзакции
func (s *PrReviewersStorage) ReassignReviewerAtomic(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	prUUID, err := uuid.Parse(prID)
	if err != nil {
		return ErrInvalidUUID
	}

	oldReviewerUUID, err := uuid.Parse(oldReviewerID)
	if err != nil {
		return ErrInvalidUUID
	}

	newReviewerUUID, err := uuid.Parse(newReviewerID)
	if err != nil {
		return ErrInvalidUUID
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Удаляем старого ревьюера
	deleteQuery := `DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2`
	_, err = tx.Exec(ctx, deleteQuery, prUUID, oldReviewerUUID)
	if err != nil {
		return err
	}

	// Добавляем нового ревьюера
	insertQuery := `
		INSERT INTO pr_reviewers (pull_request_id, reviewer_id, assigned_at)
		VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, insertQuery, prUUID, newReviewerUUID, time.Now())
	if err != nil {
		return err
	}

	// Коммитим транзакцию
	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
