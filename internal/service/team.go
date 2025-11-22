package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Estriper0/avito_intership/internal/handlers/dto"
	"github.com/Estriper0/avito_intership/internal/models"
	"github.com/Estriper0/avito_intership/internal/repository"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

type TeamService struct {
	teamRepo  repository.ITeamRepo
	userRepo  repository.IUserRepo
	trManager *manager.Manager
	logger    *slog.Logger
}

func NewTeamService(teamRepo repository.ITeamRepo, userRepo repository.IUserRepo, trManager *manager.Manager, logger *slog.Logger) *TeamService {
	return &TeamService{
		teamRepo:  teamRepo,
		userRepo:  userRepo,
		trManager: trManager,
		logger:    logger,
	}
}

func (s *TeamService) Add(ctx context.Context, team *dto.Team) (int, error) {
	var id int
	//Do everything in a transaction
	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		//Add a team to the table teams
		teamId, err := s.teamRepo.Create(ctx, &models.Team{Name: team.TeamName})
		if err != nil {
			if errors.Is(err, repository.ErrAlreadyExists) {
				return ErrTeamAlreadyExists
			}
			s.logger.Error("TeamService.Add:teamRepo.Create - Internal error", slog.String("error", err.Error()))
			return ErrInternal
		}
		id = teamId

		//Add/Update all members to the table users
		for _, user := range team.Members {
			_, err := s.userRepo.CreateOrUpdate(ctx, &models.User{
				UserId:   user.UserId,
				Username: user.Username,
				TeamId:   teamId,
				IsActive: user.IsActive,
			})
			if err != nil {
				s.logger.Error("TeamService.Add:userRepo.Create - Internal error", slog.String("error", err.Error()))
				return ErrInternal
			}
		}

		return nil
	})

	return id, err
}

func (s *TeamService) Get(ctx context.Context, teamName string) (*dto.Team, error) {
	teamId, err := s.teamRepo.GetIdByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.logger.Error("TeamService.Get:teamRepo.GetIdByName - Internal error", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	//Getting all team members
	users, err := s.userRepo.GetAllByTeam(ctx, teamId)
	if err != nil {
		s.logger.Error("TeamService.Get:userRepo.GetAllByTeam - Internal error", slog.String("error", err.Error()))
		return nil, ErrInternal
	}

	var members []dto.Members
	for _, user := range users {
		members = append(members, dto.Members{
			UserId:   user.UserId,
			Username: user.Username,
			IsActive: user.IsActive,
		})
	}

	return &dto.Team{
		TeamName: teamName,
		Members:  members,
	}, err
}

func (s *TeamService) GetStatsPR(ctx context.Context, teamName string) (*dto.TeamStatsPrResponse, error) {
	team, err := s.teamRepo.GetStatsPRByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.logger.Error("TeamService.GetStatsPR:teamRepo.GetStatsPR - Internal error", slog.String("error", err.Error()))
		return nil, ErrInternal
	}
	
	return &dto.TeamStatsPrResponse{
		Name:     team.Name,
		TotalPr:  team.TotalPr,
		OpenPr:   team.OpenPr,
		MergedPr: team.MergedPr,
	}, nil
}
