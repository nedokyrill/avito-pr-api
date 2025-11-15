package domain

import "github.com/nedokyrill/avito-pr-api/internal/generated"

const MaxReviewersCount int = 2

const (
	PullRequestStatusOPEN   PullRequestStatus = generated.PullRequestStatusOPEN
	PullRequestStatusMERGED PullRequestStatus = generated.PullRequestStatusMERGED
)

const (
	ErrCreatePRMsg         string = "error with creating pull request"
	ErrMergePRMsg          string = "error with merging pull request"
	ErrReassignReviewerMsg string = "error with reassigning reviewer"

	ErrCreateTeamMsg string = "error with creating team"
	ErrGetTeamMsg    string = "error with getting team"

	ErrSetActiveMsg           string = "error with setting active state"
	ErrGetUserReviewsMsg      string = "error with getting user reviews"
	ErrDeactivatingUsersMsg   string = "error with deactivating users"
)
