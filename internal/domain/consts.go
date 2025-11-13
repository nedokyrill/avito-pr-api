package domain

import "github.com/nedokyrill/avito-pr-api/internal/generated"

const MaxReviewersCount int = 2

const (
	PullRequestStatusOPEN   PullRequestStatus = generated.PullRequestStatusOPEN
	PullRequestStatusMERGED PullRequestStatus = generated.PullRequestStatusMERGED
)
