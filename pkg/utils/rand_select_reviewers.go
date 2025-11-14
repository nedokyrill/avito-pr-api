package utils

import (
	"math/rand"
	"time"

	"github.com/nedokyrill/avito-pr-api/internal/domain"
)

func RandSelectReviewers(members []domain.TeamMember, authorID string, maxCount int) []string {
	if maxCount <= 0 {
		return []string{}
	}

	var candidates []string
	for _, member := range members {
		if member.IsActive && member.UserId != authorID {
			candidates = append(candidates, member.UserId)
		}
	}

	if len(candidates) <= maxCount {
		return candidates
	}

	//nolint:gosec // G404: math/rand достаточно для случайного выбора ревьюеров
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	return candidates[:maxCount]
}
