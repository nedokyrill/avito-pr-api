package teamStorage

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/db"
)

var ErrNoUsersInTeam = errors.New(domain.NoUsersInTeamErr)
var ErrTeamNotExists = errors.New(domain.TeamNotExistsErr)

type TeamStorage struct {
	db db.Querier
}

func NewTeamStorage(db db.Querier) *TeamStorage {
	return &TeamStorage{
		db: db,
	}
}

func (s *TeamStorage) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	var teamID uuid.UUID
	query := `SELECT id FROM teams WHERE name = $1`
	err := s.db.QueryRow(ctx, query, teamName).Scan(&teamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTeamNotExists
		}
		return nil, err
	}

	membersQuery := `
		SELECT id, name, is_active 
		FROM users 
		WHERE team_id = $1
		ORDER BY name`

	rows, err := s.db.Query(ctx, membersQuery, teamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoUsersInTeam
		}
		return nil, err
	}
	defer rows.Close()

	var members []domain.TeamMember
	for rows.Next() {
		var userUUID uuid.UUID
		var username string
		var isActive bool

		if err = rows.Scan(&userUUID, &username, &isActive); err != nil {
			return nil, err
		}

		members = append(members, domain.TeamMember{
			UserId:   userUUID.String(),
			Username: username,
			IsActive: isActive,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	team := &domain.Team{
		TeamName: teamName,
		Members:  members,
	}

	return team, nil
}

// CreateTeamWithMembers создает команду и добавляет участников атомарно в транзакции
func (s *TeamStorage) CreateTeamWithMembers(ctx context.Context, teamName string, members []domain.TeamMember) (uuid.UUID, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Создаем команду
	var teamID uuid.UUID
	query := `INSERT INTO teams (name) VALUES ($1) RETURNING id`
	err = tx.QueryRow(ctx, query, teamName).Scan(&teamID)
	if err != nil {
		return uuid.Nil, err
	}

	// Добавляем всех участников
	userQuery := `
		INSERT INTO users (id, name, team_id, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE 
		SET name = EXCLUDED.name,
		    team_id = EXCLUDED.team_id,
		    is_active = EXCLUDED.is_active`

	for _, member := range members {
		userUUID, parseErr := uuid.Parse(member.UserId)
		if parseErr != nil {
			return uuid.Nil, parseErr
		}

		_, err = tx.Exec(ctx, userQuery, userUUID, member.Username, teamID, member.IsActive)
		if err != nil {
			return uuid.Nil, err
		}
	}

	// Коммитим транзакцию
	if err = tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}

	return teamID, nil
}
