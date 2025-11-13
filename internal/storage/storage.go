package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
)

type TeamRepositoryInterface interface {
	CreateTeam(ctx context.Context, teamName string) (uuid.UUID, error)
	GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error)
	GetTeamMembersByTeamName(ctx context.Context, teamName string) ([]*domain.TeamMember, error)
}

type UserRepositoryInterface interface {
	CreateOrUpdateUser(ctx context.Context, user *domain.User, teamID uuid.UUID) error
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
	SetUserIsActive(ctx context.Context, userID string, isActive bool) error
}

type PullRequestRepositoryInterface interface {
	CreatePullRequest(ctx context.Context, pr *domain.PullRequest) error
	GetPullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) error
}

type PrReviewersRepositoryInterface interface {
	RemoveReviewer(ctx context.Context, prID string, reviewerID string) error
	AddReviewer(ctx context.Context, prID string, reviewerID string) error
	GetAssignedReviewers(ctx context.Context, prID string) ([]string, error)
	GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
}
