package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/Estriper0/avito_intership/internal/models"
	"github.com/Estriper0/avito_intership/internal/repository"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepo struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewTeamRepo(db *pgxpool.Pool, c *trmpgx.CtxGetter) *TeamRepo {
	return &TeamRepo{
		db:     db,
		getter: c,
	}
}

func (r *TeamRepo) Create(ctx context.Context, team *models.Team) (int, error) {
	query := `
		INSERT INTO teams (name) VALUES ($1) RETURNING id
	`
	var id int

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err := conn.QueryRow(ctx, query, team.Name).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return 0, repository.ErrAlreadyExists
			}
		}
		return 0, fmt.Errorf("db:TeamRepo.Create:Exec - %s", err.Error())
	}

	return id, nil
}

func (r *TeamRepo) GetIdByName(ctx context.Context, teamName string) (int, error) {
	query := `
		SELECT id FROM teams WHERE name = $1
	`
	var id int

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err := conn.QueryRow(ctx, query, teamName).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, repository.ErrNotFound
		}
		return 0, fmt.Errorf("db:TeamRepo.GetIdByName:QueryRow - %s", err.Error())
	}

	return id, nil
}

func (r *TeamRepo) GetNameById(ctx context.Context, teamId int) (string, error) {
	query := `
		SELECT name FROM teams WHERE id = $1
	`
	var teamName string

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err := conn.QueryRow(ctx, query, teamId).Scan(&teamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", repository.ErrNotFound
		}
		return "", fmt.Errorf("db:TeamRepo.GetNameById:QueryRow - %s", err.Error())
	}

	return teamName, nil
}
