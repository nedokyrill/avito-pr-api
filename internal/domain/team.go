package domain

import "github.com/nedokyrill/avito-pr-api/internal/generated"

// Реэкспорт типов из generated для использования в доменной логике

type Team = generated.Team
type TeamMember = generated.TeamMember

type DeactivateTeamMembersRequest struct {
	TeamName string   `json:"team_name" binding:"required"`
	UserIDs  []string `json:"user_ids"` // если пустой - деактивировать всех
}
