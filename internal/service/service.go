package service

import (
	"context"

	"github.com/Estriper0/avito_intership/internal/handlers/dto"
)

type IUserService interface {
	SetIsActive(ctx context.Context, req *dto.SetIsActiveRequest) (*dto.UserResponse, error)
	GetReview(ctx context.Context, userId string) ([]dto.ReviewResponse, error)
	GetStatsReview(ctx context.Context) ([]dto.UserStatsReviewResponse, error)
}

type ITeamService interface {
	Add(ctx context.Context, team *dto.Team) (int, error)
	Get(ctx context.Context, teamName string) (*dto.Team, error)
}

type IPullRequestService interface {
	Create(ctx context.Context, pr *dto.PrCreateRequest) (*dto.PullRequest, error)
	Merge(ctx context.Context, prId string) (*dto.MergeResponse, error)
	Reassign(ctx context.Context, req *dto.ReassignRequest) (*dto.ReassignResponse, error)
}
