package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/Estriper0/avito_intership/internal/models"
	"github.com/Estriper0/avito_intership/internal/repository"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewUserRepo(db *pgxpool.Pool, c *trmpgx.CtxGetter) *UserRepo {
	return &UserRepo{
		db:     db,
		getter: c,
	}
}

func (r *UserRepo) CreateOrUpdate(ctx context.Context, user *models.User) (string, error) {
	query := `
        INSERT INTO users (user_id, username, team_id, is_active) 
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id) 
        DO UPDATE SET 
            username = EXCLUDED.username,
            team_id = EXCLUDED.team_id,
			is_active = EXCLUDED.is_active
        RETURNING user_id
    `
	var userId string

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err := conn.QueryRow(ctx, query, user.UserId, user.Username, user.TeamId, user.IsActive).Scan(&userId)
	if err != nil {
		return "", fmt.Errorf("db:UserRepo.CreateOrUpdate:QueryRow - %s", err.Error())
	}
	return userId, nil
}

func (r *UserRepo) GetAllByTeam(ctx context.Context, teamId int) ([]models.User, error) {
	query := `
		SELECT * FROM users WHERE team_id = $1
	`

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, teamId)
	if err != nil {
		return nil, fmt.Errorf("db:UserRepo.GetAllByTeam:Query - %s", err.Error())
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.UserId,
			&user.Username,
			&user.TeamId,
			&user.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("db:UserRepo.GetAllByTeam:Scan - %s", err.Error())
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db:UserRepo.GetAllByTeam:rows - %s", err.Error())
	}

	return users, nil
}

func (r *UserRepo) UpdateIsActive(ctx context.Context, userId string, isActive bool) (*models.User, error) {
	query := `
		UPDATE users SET is_active = $1 WHERE user_id = $2 RETURNING *
	`
	var user models.User

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err := conn.QueryRow(ctx, query, isActive, userId).Scan(
		&user.UserId,
		&user.Username,
		&user.TeamId,
		&user.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("db:UserRepo.UpdateIsActive:QueryRow - %s", err.Error())
	}

	return &user, nil
}

func (r *UserRepo) ExistsById(ctx context.Context, userId string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT * FROM users WHERE user_id = $1)
	`
	var exists bool

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err := conn.QueryRow(ctx, query, userId).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("db:UserRepo.ExistsById:QueryRow - %s", err.Error())
	}
	return exists, nil
}

func (r *UserRepo) GetActiveTeamMembersById(ctx context.Context, userId string) ([]models.User, error) {
	query := `
		SELECT user_id, username, team_id, is_active 
		FROM users 
		WHERE team_id = (SELECT team_id FROM users WHERE user_id = $1) AND
		is_active = true AND user_id != $1
	`

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("db:UserRepo.GetActiveTeamMembersById:Query - %s", err.Error())
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.UserId,
			&user.Username,
			&user.TeamId,
			&user.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("db:UserRepo.GetActiveTeamMembersById:Scan - %s", err.Error())
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db:UserRepo.GetActiveTeamMembersById:rows - %s", err.Error())
	}

	return users, nil
}
