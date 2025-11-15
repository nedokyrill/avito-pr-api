package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
)

type TeamRepositoryInterface interface {
	GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error)
	CreateTeamWithMembers(ctx context.Context, teamName string, members []domain.TeamMember) (uuid.UUID, error)
	DeactivateTeamMembers(ctx context.Context, teamName string, userIDs []string, reassignments []domain.ReviewerReassignment) ([]string, error)
}

type UserRepositoryInterface interface {
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
	SetUserIsActive(ctx context.Context, userID string, isActive bool) error
}

type PullRequestRepositoryInterface interface {
	GetPullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) error
	SetNeedMoreReviewers(ctx context.Context, prID string, needMore bool) error
	CreatePullRequestWithReviewers(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string, needMoreReviewers bool) error
}

type PrReviewersRepositoryInterface interface {
	GetAssignedReviewers(ctx context.Context, prID string) ([]string, error)
	GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
	ReassignReviewerAtomic(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
}
