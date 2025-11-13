package teamService

import (
	"github.com/nedokyrill/avito-pr-api/internal/storage"
)

type TeamServiceImpl struct {
	teamRepo storage.TeamRepositoryInterface
	userRepo storage.UserRepositoryInterface
}

func NewTeamService(
	teamRepo storage.TeamRepositoryInterface,
	userRepo storage.UserRepositoryInterface,
) *TeamServiceImpl {
	return &TeamServiceImpl{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}
