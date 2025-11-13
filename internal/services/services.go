package services

import (
	"github.com/gin-gonic/gin"
)

type TeamService interface {
	CreateTeam(c *gin.Context)
	GetTeam(c *gin.Context)
}

type UserService interface {
	SetIsActive(c *gin.Context)
	GetUserReviews(c *gin.Context)
}

type PullRequestService interface {
	CreatePullRequest(c *gin.Context)
	MergePullRequest(c *gin.Context)
	ReassignReviewer(c *gin.Context)
}
