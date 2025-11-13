package userStorage

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
)

var ErrInvalidUUID = errors.New(domain.InvalidUUIDErr)

type UserStorage struct {
	db *pgxpool.Pool
}

func NewUserStorage(db *pgxpool.Pool) *UserStorage {
	return &UserStorage{
		db: db,
	}
}

func (s *UserStorage) CreateOrUpdateUser(ctx context.Context, user *domain.User, teamID uuid.UUID) error {
	userUUID, err := uuid.Parse(user.UserId)
	if err != nil {
		return ErrInvalidUUID
	}

	query := `
		INSERT INTO users (id, name, team_id, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE 
		SET name = EXCLUDED.name,
		    team_id = EXCLUDED.team_id,
		    is_active = EXCLUDED.is_active`

	_, err = s.db.Exec(ctx, query, userUUID, user.Username, teamID, user.IsActive)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStorage) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	var username string
	var teamName string
	var isActive bool

	query := `
		SELECT u.name, t.name, u.is_active
		FROM users u
		LEFT JOIN teams t ON u.team_id = t.id
		WHERE u.id = $1`

	err = s.db.QueryRow(ctx, query, userUUID).Scan(&username, &teamName, &isActive)
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
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return ErrInvalidUUID
	}

	query := `
		UPDATE users 
		SET is_active = $1 
		WHERE id = $2`

	_, err = s.db.Exec(ctx, query, isActive, userUUID)
	if err != nil {
		return err
	}

	return nil
}
