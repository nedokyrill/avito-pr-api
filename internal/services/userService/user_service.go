package userService

import (
	"github.com/nedokyrill/avito-pr-api/internal/storage"
)

type UserServiceImpl struct {
	userRepo        storage.UserRepositoryInterface
	prReviewersRepo storage.PrReviewersRepositoryInterface
	teamRepo        storage.TeamRepositoryInterface
}

func NewUserService(
	userRepo storage.UserRepositoryInterface,
	prReviewersRepo storage.PrReviewersRepositoryInterface,
	teamRepo storage.TeamRepositoryInterface,
) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo:        userRepo,
		prReviewersRepo: prReviewersRepo,
		teamRepo:        teamRepo,
	}
}
