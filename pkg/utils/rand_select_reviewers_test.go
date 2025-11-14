package utils

import (
	"testing"

	"github.com/nedokyrill/avito-pr-api/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUser1  = "user1"
	testUser2  = "user2"
	testUser3  = "user3"
	testUser4  = "user4"
	testUser5  = "user5"
	testUser6  = "user6"
	testUser7  = "user7"
	testUser8  = "user8"
	testUser9  = "user9"
	testUser10 = "user10"
)

func TestRandSelectReviewers(t *testing.T) {
	t.Run("all members are active and author excluded", func(t *testing.T) {
		members := []domain.TeamMember{
			{UserId: testUser1, IsActive: true},
			{UserId: testUser2, IsActive: true},
			{UserId: testUser3, IsActive: true},
			{UserId: testUser4, IsActive: true},
		}
		authorID := testUser1
		maxCount := 2

		result := RandSelectReviewers(members, authorID, maxCount)

		require.Len(t, result, maxCount)
		assert.NotContains(t, result, authorID, "Author should not be in reviewers")

		// проверяем, что все выбранные пользователи из кандидатов
		for _, reviewer := range result {
			assert.Contains(t, []string{testUser2, testUser3, testUser4}, reviewer)
		}
	})

	t.Run("filters out inactive members", func(t *testing.T) {
		members := []domain.TeamMember{
			{UserId: testUser1, IsActive: true},
			{UserId: testUser2, IsActive: false},
			{UserId: testUser3, IsActive: true},
			{UserId: testUser4, IsActive: false},
		}
		authorID := testUser5
		maxCount := 5

		result := RandSelectReviewers(members, authorID, maxCount)

		require.Len(t, result, 2, "Should only include active members")
		assert.Contains(t, result, testUser1)
		assert.Contains(t, result, testUser3)
		assert.NotContains(t, result, testUser2)
		assert.NotContains(t, result, testUser4)
	})

	t.Run("excludes author even if active", func(t *testing.T) {
		members := []domain.TeamMember{
			{UserId: testUser1, IsActive: true},
			{UserId: testUser2, IsActive: true},
			{UserId: testUser3, IsActive: true},
		}
		authorID := testUser2
		maxCount := 5

		result := RandSelectReviewers(members, authorID, maxCount)

		require.Len(t, result, 2)
		assert.NotContains(t, result, authorID)
		assert.Contains(t, result, testUser1)
		assert.Contains(t, result, testUser3)
	})

	t.Run("returns all candidates when count less than maxCount", func(t *testing.T) {
		members := []domain.TeamMember{
			{UserId: testUser1, IsActive: true},
			{UserId: testUser2, IsActive: true},
		}
		authorID := testUser3
		maxCount := 5

		result := RandSelectReviewers(members, authorID, maxCount)

		require.Len(t, result, 2, "Should return all available candidates")
		assert.Contains(t, result, testUser1)
		assert.Contains(t, result, testUser2)
	})

	t.Run("returns all candidates when count equals maxCount", func(t *testing.T) {
		members := []domain.TeamMember{
			{UserId: testUser1, IsActive: true},
			{UserId: testUser2, IsActive: true},
			{UserId: testUser3, IsActive: true},
		}
		authorID := testUser4
		maxCount := 3

		result := RandSelectReviewers(members, authorID, maxCount)

		require.Len(t, result, 3)
		assert.Contains(t, result, testUser1)
		assert.Contains(t, result, testUser2)
		assert.Contains(t, result, testUser3)
	})

	t.Run("empty members list", func(t *testing.T) {
		members := []domain.TeamMember{}
		authorID := testUser1
		maxCount := 2

		result := RandSelectReviewers(members, authorID, maxCount)

		assert.Empty(t, result)
	})

	t.Run("nil members list", func(t *testing.T) {
		var members []domain.TeamMember
		authorID := testUser1
		maxCount := 2

		result := RandSelectReviewers(members, authorID, maxCount)

		assert.Empty(t, result)
	})

	t.Run("only author in members", func(t *testing.T) {
		members := []domain.TeamMember{
			{UserId: testUser1, IsActive: true},
		}
		authorID := testUser1
		maxCount := 2

		result := RandSelectReviewers(members, authorID, maxCount)

		assert.Empty(t, result, "Should return empty when only author is available")
	})

	t.Run("only inactive members", func(t *testing.T) {
		members := []domain.TeamMember{
			{UserId: testUser1, IsActive: false},
			{UserId: testUser2, IsActive: false},
			{UserId: testUser3, IsActive: false},
		}
		authorID := testUser4
		maxCount := 2

		result := RandSelectReviewers(members, authorID, maxCount)

		assert.Empty(t, result)
	})

	t.Run("maxCount is zero", func(t *testing.T) {
		members := []domain.TeamMember{
			{UserId: testUser1, IsActive: true},
			{UserId: testUser2, IsActive: true},
			{UserId: testUser3, IsActive: true},
		}
		authorID := testUser4
		maxCount := 0

		result := RandSelectReviewers(members, authorID, maxCount)

		assert.Empty(t, result)
	})

	t.Run("maxCount is negative", func(t *testing.T) {
		members := []domain.TeamMember{
			{UserId: testUser1, IsActive: true},
			{UserId: testUser2, IsActive: true},
		}
		authorID := testUser3
		maxCount := -1

		result := RandSelectReviewers(members, authorID, maxCount)

		assert.Empty(t, result)
	})

	t.Run("large team with maxCount selection", func(t *testing.T) {
		members := []domain.TeamMember{
			{UserId: testUser1, IsActive: true},
			{UserId: testUser2, IsActive: true},
			{UserId: testUser3, IsActive: true},
			{UserId: testUser4, IsActive: true},
			{UserId: testUser5, IsActive: true},
			{UserId: testUser6, IsActive: true},
			{UserId: testUser7, IsActive: true},
			{UserId: testUser8, IsActive: true},
			{UserId: testUser9, IsActive: true},
			{UserId: testUser10, IsActive: true},
		}
		authorID := testUser1
		maxCount := 3

		result := RandSelectReviewers(members, authorID, maxCount)

		require.Len(t, result, maxCount, "Should select exactly maxCount reviewers")
		assert.NotContains(t, result, authorID)

		// проверяем уникальность
		unique := make(map[string]bool)
		for _, reviewer := range result {
			assert.False(t, unique[reviewer], "Reviewers should be unique")
			unique[reviewer] = true
		}
	})

	t.Run("mixed active and inactive with author", func(t *testing.T) {
		members := []domain.TeamMember{
			{UserId: testUser1, IsActive: true},
			{UserId: testUser2, IsActive: false},
			{UserId: testUser3, IsActive: true},
			{UserId: testUser4, IsActive: false},
			{UserId: testUser5, IsActive: true},
			{UserId: testUser6, IsActive: true},
		}
		authorID := testUser3
		maxCount := 2

		result := RandSelectReviewers(members, authorID, maxCount)

		require.Len(t, result, maxCount)
		assert.NotContains(t, result, authorID)
		assert.NotContains(t, result, testUser2)
		assert.NotContains(t, result, testUser4)

		// все выбранные должны быть из активных (кроме автора)
		validCandidates := []string{testUser1, testUser5, testUser6}
		for _, reviewer := range result {
			assert.Contains(t, validCandidates, reviewer)
		}
	})

	t.Run("randomness - multiple calls produce different results", func(t *testing.T) {
		members := []domain.TeamMember{
			{UserId: testUser1, IsActive: true},
			{UserId: testUser2, IsActive: true},
			{UserId: testUser3, IsActive: true},
			{UserId: testUser4, IsActive: true},
			{UserId: testUser5, IsActive: true},
		}
		authorID := testUser6
		maxCount := 2

		// выполняем несколько раз и проверяем, что хотя бы раз результат отличается
		results := make(map[string]bool)
		for i := 0; i < 20; i++ {
			result := RandSelectReviewers(members, authorID, maxCount)
			require.Len(t, result, maxCount)
			key := result[0] + "," + result[1]
			results[key] = true
		}

		// должно быть хотя бы 2 разных комбинации из 20 попыток
		assert.GreaterOrEqual(t, len(results), 2, "Multiple calls should produce varied results due to randomness")
	})
}
