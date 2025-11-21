package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Estriper0/avito_intership/internal/handlers/dto"
	"github.com/Estriper0/avito_intership/internal/repository"
)

type UserService struct {
	userRepo repository.IUserRepo
	teamRepo repository.ITeamRepo
	prRepo   repository.IPullRequestRepo
	logger   *slog.Logger
}

func NewUserService(userRepo repository.IUserRepo, teamRepo repository.ITeamRepo, prRepo repository.IPullRequestRepo, logger *slog.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		teamRepo: teamRepo,
		prRepo:   prRepo,
		logger:   logger,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, req *dto.SetIsActiveRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.UpdateIsActive(ctx, req.UserId, req.IsActive)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.logger.Error("UserService.SetIsActive:userRepo.UpdateIsActive - Internal error", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	teamName, err := s.teamRepo.GetNameById(ctx, user.TeamId)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.logger.Error("UserService.SetIsActive:teamRepo.GetNameById - Internal error", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	return &dto.UserResponse{
		UserId:   user.UserId,
		Username: user.Username,
		TeamName: teamName,
		IsActive: user.IsActive,
	}, err
}

func (s *UserService) GetReview(ctx context.Context, userId string) ([]dto.ReviewResponse, error) {
	pr, err := s.prRepo.GetAllReviewByUserId(ctx, userId)
	if err != nil {
		s.logger.Error("UserService.GetReview:prRepo.GetAllReviewByUserId - Internal error", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	var reviews []dto.ReviewResponse
	for _, p := range pr {
		//Getting a status name
		status, err := s.prRepo.GetStatusById(ctx, p.StatusId)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, ErrNotFound
			}
			s.logger.Error("UserService.GetReview:prRepo.GetStatusById - Internal error", slog.String("error", err.Error()))
			return nil, ErrInternal
		}

		reviews = append(reviews, dto.ReviewResponse{
			PrId:     p.PrId,
			PrName:   p.Name,
			AuthorId: p.AuthorId,
			Status:   status,
		})

	}
	return reviews, nil
}

func (s *UserService) GetStatsReview(ctx context.Context) ([]dto.UserStatsReviewResponse, error) {
	users, err := s.userRepo.GetStatsReview(ctx)
	if err != nil {
		s.logger.Error("UserService.GetStatsReview:userRepo.GetStatsReview - Internal error", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	var resp []dto.UserStatsReviewResponse
	for _, user := range users {
		resp = append(resp, dto.UserStatsReviewResponse{
			UserId:          user.UserId,
			Username:        user.Username,
			CountOpenReview: user.CountOpenReview,
		})

	}
	return resp, nil
}
