package domain

import "github.com/nedokyrill/avito-pr-api/internal/generated"

// Реэкспорт типов из generated для использования в доменной логике

type PullRequest = generated.PullRequest

type PullRequestShort = generated.PullRequestShort

type PullRequestStatus = generated.PullRequestStatus

const (
	PullRequestStatusOPEN   PullRequestStatus = generated.PullRequestStatusOPEN
	PullRequestStatusMERGED PullRequestStatus = generated.PullRequestStatusMERGED
)
