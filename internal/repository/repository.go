package repository

import (
	"context"

	"github.com/Estriper0/avito_intership/internal/models"
)

type IUserRepo interface {
	CreateOrUpdate(ctx context.Context, user *models.User) (string, error)
	GetAllByTeam(ctx context.Context, teamId int) ([]models.User, error)
	UpdateIsActive(ctx context.Context, userId string, isActive bool) (*models.User, error)
	ExistsById(ctx context.Context, userId string) (bool, error)
	GetActiveTeamMembersById(ctx context.Context, userId string) ([]models.User, error)
	GetStatsReview(ctx context.Context) ([]models.UserStatsReview, error)
}

type ITeamRepo interface {
	Create(ctx context.Context, team *models.Team) (int, error)
	GetIdByName(ctx context.Context, teamName string) (int, error)
	GetNameById(ctx context.Context, teamId int) (string, error)
}

type IPullRequestRepo interface {
	Create(ctx context.Context, pr *models.PullRequest) (*models.PullRequest, error)
	AddReviewers(ctx context.Context, prId string, reviewersId []string) error
	GetReviewers(ctx context.Context, prId string) ([]string, error)
	Merge(ctx context.Context, prId string) (*models.PullRequest, error)
	GetAllReviewByUserId(ctx context.Context, userId string) ([]models.PullRequest, error)
	GetStatusById(ctx context.Context, statusId int) (string, error)
	GetById(ctx context.Context, prId string) (*models.PullRequest, error)
	UpdateReviewer(ctx context.Context, prId string, oldReviewer string, newReviewer string) (string, error)
}
