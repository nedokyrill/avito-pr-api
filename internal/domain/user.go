package domain

import "github.com/nedokyrill/avito-pr-api/internal/generated"

// Реэкспорт типов из generated для использования в доменной логике

type User = generated.User

type SetIsActiveRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive bool   `json:"is_active"`
}

type DeactivateTeamMembersResponse struct {
	DeactivatedUserIDs []string               `json:"deactivated_user_ids"`
	Reassignments      []ReviewerReassignment `json:"reassignments"`
}

type ReviewerReassignment struct {
	PrID          string `json:"pr_id"`
	OldReviewerID string `json:"old_reviewer_id"`
	NewReviewerID string `json:"new_reviewer_id,omitempty"` // пустой = удалить без замены
}
