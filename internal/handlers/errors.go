package handlers

import (
	"github.com/gin-gonic/gin"
)

const (
	ErrStatusTeamExists  = "TEAM_EXISTS"
	ErrStatusPrExists    = "PR_EXISTS"
	ErrStatusPrMerged    = "PR_MERGED"
	ErrStatusNotAssigned = "NOT_ASSIGNED"
	ErrStatusNoCandidate = "NO_CANDIDATE"
	ErrStatusNotFound    = "NOT_FOUND"
	ErrStatusInternal    = "INTERNAL"
	ErrStatusBadRequest  = "BAD_REQUEST"
)

func respondWithError(c *gin.Context, code int, errStatus string, err error) {
	c.JSON(
		code,
		gin.H{
			"error": gin.H{
				"code":    errStatus,
				"message": err.Error(),
			},
		},
	)
}
