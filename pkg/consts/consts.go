package consts

import "time"

const (
	PgxTimeout = 5 * time.Second
	GsTimeout  = 5 * time.Second
)

const (
	InvalidUUIDErr string = "invalid uuid"
	InvalidPRIDErr string = "invalid pull request id"

	TeamNotExistsErr string = "team does not exist"
	NoUsersInTeamErr string = "no users in team"
)
