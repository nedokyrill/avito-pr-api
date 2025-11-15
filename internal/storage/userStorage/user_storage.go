package userStorage

import (
	"context"

	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/db"
)

type UserStorage struct {
	db db.Querier
}

func NewUserStorage(db db.Querier) *UserStorage {
	return &UserStorage{
		db: db,
	}
}

func (s *UserStorage) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	var username string
	var teamName string
	var isActive bool

	query := `
		SELECT u.name, t.name, u.is_active
		FROM users u
		LEFT JOIN teams t ON u.team_id = t.id
		WHERE u.id = $1`

	err := s.db.QueryRow(ctx, query, userID).Scan(&username, &teamName, &isActive)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		UserId:   userID,
		Username: username,
		TeamName: teamName,
		IsActive: isActive,
	}

	return user, nil
}

func (s *UserStorage) SetUserIsActive(ctx context.Context, userID string, isActive bool) error {
	query := `
		UPDATE users 
		SET is_active = $1 
		WHERE id = $2`

	_, err := s.db.Exec(ctx, query, isActive, userID)
	if err != nil {
		return err
	}

	return nil
}
