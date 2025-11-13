package pullRequestStorage

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/pkg/consts"
)

var ErrInvalidUUID = errors.New(consts.InvalidUUIDErr)
var ErrInvalidPRID = errors.New(consts.InvalidPRIDErr)

type PullRequestStorage struct {
	db *pgxpool.Pool
}

func NewPullRequestStorage(db *pgxpool.Pool) *PullRequestStorage {
	return &PullRequestStorage{
		db: db,
	}
}

func (s *PullRequestStorage) CreatePullRequest(ctx context.Context, pr *domain.PullRequest) error {
	authorUUID, err := uuid.Parse(pr.AuthorId)
	if err != nil {
		return ErrInvalidUUID
	}

	query := `INSERT INTO pull_requests (name, author_id, status) VALUES ($1, $2, $3)`

	_, err = s.db.Exec(ctx, query, pr.PullRequestName, authorUUID, string(pr.Status))
	if err != nil {
		return err
	}

	return nil
}

func (s *PullRequestStorage) GetPullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	prIDInt, err := strconv.Atoi(prID)
	if err != nil {
		return nil, ErrInvalidPRID
	}

	var id int
	var name string
	var authorUUID uuid.UUID
	var status string
	var createdAt time.Time
	var mergedAt *time.Time

	query := `
		SELECT id, name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE id = $1`

	err = s.db.QueryRow(ctx, query, prIDInt).Scan(&id, &name, &authorUUID, &status, &createdAt, &mergedAt)
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
	prIDInt, err := strconv.Atoi(prID)
	if err != nil {
		return ErrInvalidPRID
	}

	mergedAt := time.Now()

	query := `
		UPDATE pull_requests
		SET status = $1, merged_at = $2
		WHERE id = $3 AND status != $1`

	_, err = s.db.Exec(ctx, query, string(domain.PullRequestStatusMERGED), mergedAt, prIDInt)
	if err != nil {
		return err
	}

	return nil
}

