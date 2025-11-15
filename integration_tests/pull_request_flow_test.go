//go:build integration

package integration_tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

const ConnStr string = "postgres://postgres:123123123@localhost:5432/postgres?sslmode=disable"

func TestIntegrationPullRequestFlow(t *testing.T) {
	ctx := context.Background()

	conn, err := pgxpool.New(ctx, ConnStr)
	require.NoError(t, err, "не удалось подключиться к бд")
	defer conn.Close()

	err = conn.Ping(ctx)
	require.NoError(t, err, "не удалось выполнить пинг бд")

	// Создание команды
	teamName := "Backend"
	sqlInsertTeam := `INSERT INTO teams (name) VALUES ($1) RETURNING id`
	var teamID uuid.UUID
	err = conn.QueryRow(ctx, sqlInsertTeam, teamName).Scan(&teamID)
	require.NoError(t, err, "команда не создалась")
	require.NotEqual(t, uuid.Nil, teamID, "ID команды должен быть сгенерирован")

	// Создание пользователей (ID - varchar, генерируем вручную)
	users := []struct {
		name     string
		isActive bool
		id       string
	}{
		{name: "Alice", id: "user-alice-1", isActive: true},     // автор
		{name: "Bob", id: "user-bob-1", isActive: true},         // ревьювер 1
		{name: "Charlie", id: "user-charlie-1", isActive: true}, // ревьювер 2
		{name: "David", id: "user-david-1", isActive: false},    // неактивный (не должен быть назначен)
	}

	sqlInsertUser := `INSERT INTO users (id, name, team_id, is_active) VALUES ($1, $2, $3, $4)`
	for i := range users {
		_, err = conn.Exec(ctx, sqlInsertUser, users[i].id, users[i].name, teamID, users[i].isActive)
		require.NoError(t, err, "пользователь %s не создался", users[i].name)
		require.NotEmpty(t, users[i].id, "ID пользователя должен быть задан")
	}

	authorID := users[0].id
	reviewer1ID := users[1].id
	reviewer2ID := users[2].id

	// Создание Pull Request
	prName := "Add new feature"
	prID := "pr-feature-1"
	sqlInsertPR := `INSERT INTO pull_requests (id, name, author_id, status) VALUES ($1, $2, $3, 'OPEN') RETURNING id, created_at`
	var returnedPrID string
	var createdAt time.Time
	err = conn.QueryRow(ctx, sqlInsertPR, prID, prName, authorID).Scan(&returnedPrID, &createdAt)
	require.NoError(t, err, "PR не создался")
	require.Equal(t, prID, returnedPrID, "ID PR должен совпадать")
	require.False(t, createdAt.IsZero(), "created_at должен быть заполнен")

	// Назначение ревьюеров (до 2 активных участников, исключая автора)
	sqlInsertReviewer := `INSERT INTO pr_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)`
	_, err = conn.Exec(ctx, sqlInsertReviewer, prID, reviewer1ID)
	require.NoError(t, err, "ревьювер 1 не добавился")

	_, err = conn.Exec(ctx, sqlInsertReviewer, prID, reviewer2ID)
	require.NoError(t, err, "ревьювер 2 не добавился")

	// Проверка, что ревьюверы назначены
	sqlCountReviewers := `SELECT COUNT(*) FROM pr_reviewers WHERE pull_request_id = $1`
	var reviewersCount int
	err = conn.QueryRow(ctx, sqlCountReviewers, prID).Scan(&reviewersCount)
	require.NoError(t, err, "не удалось получить количество ревьюеров")
	require.Equal(t, 2, reviewersCount, "должно быть назначено 2 ревьювера")

	// Проверка, что автор НЕ назначен ревьювером
	sqlCheckAuthorNotReviewer := `SELECT COUNT(*) FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2`
	var authorAsReviewerCount int
	err = conn.QueryRow(ctx, sqlCheckAuthorNotReviewer, prID, authorID).Scan(&authorAsReviewerCount)
	require.NoError(t, err, "не удалось проверить автора")
	require.Equal(t, 0, authorAsReviewerCount, "автор не должен быть назначен ревьювером")

	// Проверка, что неактивный пользователь НЕ назначен
	inactiveUserID := users[3].id
	sqlCheckInactiveNotReviewer := `SELECT COUNT(*) FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2`
	var inactiveAsReviewerCount int
	err = conn.QueryRow(ctx, sqlCheckInactiveNotReviewer, prID, inactiveUserID).Scan(&inactiveAsReviewerCount)
	require.NoError(t, err, "не удалось проверка неактивного пользователя")
	require.Equal(t, 0, inactiveAsReviewerCount, "неактивный пользователь не должен быть назначен ревьювером")

	// Получение списка назначенных ревьюеров
	sqlGetReviewers := `SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id = $1 ORDER BY reviewer_id`
	rows, err := conn.Query(ctx, sqlGetReviewers, prID)
	require.NoError(t, err, "не удалось выполнить запрос ревьюеров")
	defer rows.Close()

	var assignedReviewers []string
	for rows.Next() {
		var reviewerID string
		err = rows.Scan(&reviewerID)
		require.NoError(t, err, "не удалось прочитать ID ревьювера")
		assignedReviewers = append(assignedReviewers, reviewerID)
	}
	require.Len(t, assignedReviewers, 2, "должно быть 2 ревьювера")
	require.Contains(t, assignedReviewers, reviewer1ID, "Bob должен быть в списке ревьюеров")
	require.Contains(t, assignedReviewers, reviewer2ID, "Charlie должен быть в списке ревьюеров")

	// Проверка статуса PR до мерджа
	sqlCheckPRStatus := `SELECT status, merged_at FROM pull_requests WHERE id = $1`
	var status string
	var mergedAt *time.Time
	err = conn.QueryRow(ctx, sqlCheckPRStatus, prID).Scan(&status, &mergedAt)
	require.NoError(t, err, "не удалось выполнить запрос статуса PR")
	require.Equal(t, "OPEN", status, "статус должен быть OPEN")
	require.Nil(t, mergedAt, "merged_at должен быть NULL для открытого PR")

	// Мердж Pull Request
	sqlMergePR := `UPDATE pull_requests SET status = 'MERGED', merged_at = NOW() WHERE id = $1`
	_, err = conn.Exec(ctx, sqlMergePR, prID)
	require.NoError(t, err, "PR должен быть смержен")

	// Проверка финального состояния
	err = conn.QueryRow(ctx, sqlCheckPRStatus, prID).Scan(&status, &mergedAt)
	require.NoError(t, err, "не удалось выполнить запрос финального статуса")
	require.Equal(t, "MERGED", status, "статус должен быть MERGED")
	require.NotNil(t, mergedAt, "merged_at должен быть заполнен")
	require.True(t, mergedAt.After(createdAt), "merged_at должен быть позже created_at")

	// Проверка идемпотентности (повторный мердж)
	// Повторный update на уже смердженном PR должен просто обновить запись
	sqlMergePRAgain := `UPDATE pull_requests SET status = 'MERGED' WHERE id = $1 AND status = 'MERGED'`
	cmdTag, err := conn.Exec(ctx, sqlMergePRAgain, prID)
	require.NoError(t, err, "повторный мердж не должен вызывать ошибку")
	require.Equal(t, int64(1), cmdTag.RowsAffected(), "должна быть обновлена 1 строка (идемпотентность)")

	// Тест cascade delete (удаление команды удаляет все связанные данные)
	sqlDeleteTeam := `DELETE FROM teams WHERE id = $1`
	_, err = conn.Exec(ctx, sqlDeleteTeam, teamID)
	require.NoError(t, err, "команда должна быть удалена")

	// Проверка, что пользователи удалены
	sqlCountUsers := `SELECT COUNT(*) FROM users WHERE team_id = $1`
	var usersCount int
	err = conn.QueryRow(ctx, sqlCountUsers, teamID).Scan(&usersCount)
	require.NoError(t, err, "не удалось выполнить запрос пользователей")
	require.Equal(t, 0, usersCount, "пользователи должны быть удалены через cascade")

	// Проверка, что PR удален
	sqlCountPRs := `SELECT COUNT(*) FROM pull_requests WHERE id = $1`
	var prsCount int
	err = conn.QueryRow(ctx, sqlCountPRs, prID).Scan(&prsCount)
	require.NoError(t, err, "должен быть выполнен запрос PR")
	require.Equal(t, 0, prsCount, "PR должен быть удален через cascade")
}
