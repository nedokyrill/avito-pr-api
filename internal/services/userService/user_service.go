package userService

import (
	"github.com/nedokyrill/avito-pr-api/internal/storage"
)

type UserServiceImpl struct {
	userRepo        storage.UserRepositoryInterface
	prReviewersRepo storage.PrReviewersRepositoryInterface
}

func NewUserService(
	userRepo storage.UserRepositoryInterface,
	prReviewersRepo storage.PrReviewersRepositoryInterface,
) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo:        userRepo,
		prReviewersRepo: prReviewersRepo,
	}
}
