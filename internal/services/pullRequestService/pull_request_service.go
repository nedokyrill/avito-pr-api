package pullRequestService

import (
	"github.com/nedokyrill/avito-pr-api/internal/storage"
)

type PullRequestServiceImpl struct {
	prRepo          storage.PullRequestRepositoryInterface
	prReviewersRepo storage.PrReviewersRepositoryInterface
	userRepo        storage.UserRepositoryInterface
	teamRepo        storage.TeamRepositoryInterface
}

func NewPullRequestService(
	prRepo storage.PullRequestRepositoryInterface,
	prReviewersRepo storage.PrReviewersRepositoryInterface,
	userRepo storage.UserRepositoryInterface,
	teamRepo storage.TeamRepositoryInterface,
) *PullRequestServiceImpl {
	return &PullRequestServiceImpl{
		prRepo:          prRepo,
		prReviewersRepo: prReviewersRepo,
		userRepo:        userRepo,
		teamRepo:        teamRepo,
	}
}
