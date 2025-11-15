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
		var userID string
		var username string
		var isActive bool

		if err = rows.Scan(&userID, &username, &isActive); err != nil {
			return nil, err
		}

		members = append(members, domain.TeamMember{
			UserId:   userID,
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

func (s *TeamStorage) CreateTeamWithMembers(
	ctx context.Context,
	teamName string,
	members []domain.TeamMember,
) (uuid.UUID, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var teamID uuid.UUID
	query := `INSERT INTO teams (name) VALUES ($1) RETURNING id`
	err = tx.QueryRow(ctx, query, teamName).Scan(&teamID)
	if err != nil {
		return uuid.Nil, err
	}

	userQuery := `
		INSERT INTO users (id, name, team_id, is_active)
		VALUES ($1, $2, $3, $4)`

	for _, member := range members {
		_, err = tx.Exec(ctx, userQuery, member.UserId, member.Username, teamID, member.IsActive)
		if err != nil {
			return uuid.Nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}

	return teamID, nil
}

func (s *TeamStorage) DeactivateTeamMembers(
	ctx context.Context,
	teamName string,
	userIDs []string,
	reassignments []domain.ReviewerReassignment,
) ([]string, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Деактивируем пользователей
	query := `
		UPDATE users u
		SET is_active = false
		FROM teams t
		WHERE u.team_id = t.id 
		  AND t.name = $1`

	var args []interface{}
	args = append(args, teamName)

	if len(userIDs) > 0 {
		query += ` AND u.id = ANY($2)`
		args = append(args, userIDs)
	}

	query += ` RETURNING u.id`

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	var deactivatedIDs []string
	for rows.Next() {
		var userID string
		if err = rows.Scan(&userID); err != nil {
			rows.Close()
			return nil, err
		}
		deactivatedIDs = append(deactivatedIDs, userID)
	}
	rows.Close()

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Переназначаем ревьюверов
	for _, reassignment := range reassignments {
		deleteQuery := `DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2`
		_, err = tx.Exec(ctx, deleteQuery, reassignment.PrID, reassignment.OldReviewerID)
		if err != nil {
			return nil, err
		}

		if reassignment.NewReviewerID != "" {
			insertQuery := `
				INSERT INTO pr_reviewers (pull_request_id, reviewer_id, assigned_at)
				VALUES ($1, $2)`
			_, err = tx.Exec(ctx, insertQuery, reassignment.PrID, reassignment.NewReviewerID)
			if err != nil {
				return nil, err
			}
		}
	}

	// 3. Коммитим транзакцию
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return deactivatedIDs, nil
}
