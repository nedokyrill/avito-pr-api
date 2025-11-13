package teamStorage

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/nedokyrill/avito-pr-api/pkg/consts"
)

var ErrNoUsersInTeam = errors.New(consts.NoUsersInTeamErr)
var ErrTeamNotExists = errors.New(consts.TeamNotExistsErr)

type TeamStorage struct {
	db *pgxpool.Pool
}

func NewTeamStorage(db *pgxpool.Pool) *TeamStorage {
	return &TeamStorage{
		db: db,
	}
}

func (s *TeamStorage) CreateTeam(ctx context.Context, teamName string) (uuid.UUID, error) {
	var teamID uuid.UUID
	query := `INSERT INTO teams (name) VALUES ($1) RETURNING id`

	err := s.db.QueryRow(ctx, query, teamName).Scan(&teamID)
	if err != nil {
		return uuid.Nil, err
	}

	return teamID, nil
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

func (s *TeamStorage) GetTeamMembersByTeamName(ctx context.Context, teamName string) ([]*domain.TeamMember, error) {
	query := `
		SELECT u.id, u.name, u.is_active
		FROM users u
		JOIN teams t ON u.team_id = t.id
		WHERE t.name = $1
		ORDER BY u.name`

	rows, err := s.db.Query(ctx, query, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.TeamMember
	for rows.Next() {
		var userUUID uuid.UUID
		var username string
		var isActive bool

		if err = rows.Scan(&userUUID, &username, &isActive); err != nil {
			return nil, err
		}

		members = append(members, &domain.TeamMember{
			UserId:   userUUID.String(),
			Username: username,
			IsActive: isActive,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}
