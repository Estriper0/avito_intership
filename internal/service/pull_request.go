package service

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"

	"github.com/Estriper0/avito_intership/internal/handlers/dto"
	"github.com/Estriper0/avito_intership/internal/models"
	"github.com/Estriper0/avito_intership/internal/repository"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

type PullRequestService struct {
	prRepo    repository.IPullRequestRepo
	userRepo  repository.IUserRepo
	trManager *manager.Manager
	logger    *slog.Logger
}

func NewPullRequestService(prRepo repository.IPullRequestRepo, userRepo repository.IUserRepo, trManager *manager.Manager, logger *slog.Logger) *PullRequestService {
	return &PullRequestService{
		prRepo:    prRepo,
		userRepo:  userRepo,
		trManager: trManager,
		logger:    logger,
	}
}

func (s *PullRequestService) Create(ctx context.Context, pr *dto.PrCreateRequest) (*dto.PullRequest, error) {
	var resp *dto.PullRequest
	//Do everything in a transaction
	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		exists, err := s.userRepo.ExistsById(ctx, pr.AuthorId)
		if err != nil {
			s.logger.Error("PullRequestService.Create:userRepo.ExistsById - Internal error", slog.String("error", err.Error()))
			return ErrInternal
		} else if !exists {
			return ErrNotFound
		}

		//Getting all active users from a user's team without a user
		activeUsers, err := s.userRepo.GetActiveTeamMembersById(ctx, pr.AuthorId)
		if err != nil {
			s.logger.Error("PullRequestService.Create:userRepo.GetActiveTeamMembersById - Internal error", slog.String("error", err.Error()))
			return ErrInternal
		}

		p, err := s.prRepo.Create(ctx, &models.PullRequest{
			PrId:     pr.PrId,
			Name:     pr.PrName,
			AuthorId: pr.AuthorId,
		})
		if err != nil {
			if errors.Is(err, repository.ErrAlreadyExists) {
				return ErrPullRequestALreadyExists
			}
			s.logger.Error("PullRequestService.Create:prRepo.Create - Internal error", slog.String("error", err.Error()))
			return ErrInternal
		}

		//Getting a status name
		status, err := s.prRepo.GetStatusById(ctx, p.StatusId)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrNotFound
			}
			s.logger.Error("PullRequestService.Create:prRepo.GetStatusById - Internal error", slog.String("error", err.Error()))
			return ErrInternal
		}

		//How many users will we assign as reviewers
		var reviewersId []string
		var count int
		if len(activeUsers) > 2 {
			count = 2
		} else {
			count = len(activeUsers)
		}

		//Randomly shuffle the slice of active users and assign reviewers
		perm := rand.Perm(len(activeUsers))
		for i := 0; i < count; i++ {
			reviewersId = append(reviewersId, activeUsers[perm[i]].UserId)
		}

		err = s.prRepo.AddReviewers(ctx, pr.PrId, reviewersId)
		if err != nil {
			if errors.Is(err, repository.ErrAlreadyExists) {
				return ErrPullRequestALreadyExists
			}
			s.logger.Error("PullRequestService.Create:prRepo.AddReviewers - Internal error", slog.String("error", err.Error()))
			return ErrInternal
		}

		resp = &dto.PullRequest{
			PrId:              p.PrId,
			PrName:            p.Name,
			AuthorId:          p.AuthorId,
			Status:            status,
			AssignedReviewers: reviewersId,
		}

		return nil
	})
	return resp, err
}

func (s *PullRequestService) Merge(ctx context.Context, prId string) (*dto.MergeResponse, error) {
	pr, err := s.prRepo.Merge(ctx, prId)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.logger.Error("PullRequestService.Merge:prRepo.Merge - Internal error", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	reviewersId, err := s.prRepo.GetReviewers(ctx, prId)
	if err != nil {
		s.logger.Error("PullRequestService.Merge:prRepo.GetReviewers - Internal error", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	//Getting a status name
	status, err := s.prRepo.GetStatusById(ctx, pr.StatusId)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.logger.Error("PullRequestService.Merge:prRepo.GetStatusById - Internal error", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	return &dto.MergeResponse{
		PrId:              pr.PrId,
		PrName:            pr.Name,
		AuthorId:          pr.AuthorId,
		Status:            status,
		AssignedReviewers: reviewersId,
		MergedAt:          pr.MergedAt,
	}, nil
}

func (s *PullRequestService) Reassign(ctx context.Context, req *dto.ReassignRequest) (*dto.ReassignResponse, error) {
	pr, err := s.prRepo.GetById(ctx, req.PrId)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.logger.Error("PullRequestService.Reassign:prRepo.GetAuthorById - Internal error", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	//Getting a status name
	status, err := s.prRepo.GetStatusById(ctx, pr.StatusId)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.logger.Error("PullRequestService.Reassign:prRepo.GetStatusById - Internal error", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	if status == "MERGED" {
		return nil, ErrPullRequestMerged
	}

	reviewers, err := s.prRepo.GetReviewers(ctx, req.PrId)
	if err != nil {
		s.logger.Error("PullRequestService.Reassign:prRepo.GetReviewers - Internal error", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	//Check that the user is assigned as a reviewer
	var findReviwer bool
	for _, reviewer := range reviewers {
		if reviewer == req.OldReviewerId {
			findReviwer = true
			break
		}
	}

	if !findReviwer {
		return nil, ErrNotFound
	}

	//Getting all active users from a user's team without a author
	activeUsers, err := s.userRepo.GetActiveTeamMembersById(ctx, pr.AuthorId)
	if err != nil {
		return nil, err
	}

	//Create a candidate slice based on active users, which does not include already assigned reviewers.
	candidate := make([]string, 0, len(activeUsers))
	for _, user := range activeUsers {
		alreadyReviewer := false
		for _, reviewer := range reviewers {
			if user.UserId == reviewer {
				alreadyReviewer = true
				break
			}
		}
		if !alreadyReviewer {
			candidate = append(candidate, user.UserId)
		}
	}

	if len(candidate) == 0 {
		return nil, ErrNoCandidate
	}

	newReviewerId, err := s.prRepo.UpdateReviewer(ctx, req.PrId, req.OldReviewerId, candidate[rand.Intn(len(candidate))])
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.logger.Error("PullRequestService.Reassign:prRepo.UpdateReviewer - Internal error", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	//Replacing the old reviewer with a new.
	for i, reviewer := range reviewers {
		if reviewer == req.OldReviewerId {
			reviewers[i] = newReviewerId
			break
		}
	}

	return &dto.ReassignResponse{
		PR: &dto.PullRequest{
			PrId:              pr.PrId,
			PrName:            pr.Name,
			AuthorId:          pr.AuthorId,
			Status:            status,
			AssignedReviewers: reviewers,
		},
		NewReviewerId: newReviewerId,
	}, nil
}
