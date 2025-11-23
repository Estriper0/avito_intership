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

type PullRequestRepo struct {
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewPullRequestRepo(db *pgxpool.Pool, c *trmpgx.CtxGetter) *PullRequestRepo {
	return &PullRequestRepo{
		db:     db,
		getter: c,
	}
}

func (r *PullRequestRepo) Create(ctx context.Context, pr *models.PullRequest) (*models.PullRequest, error) {
	query := `
		INSERT INTO pull_requests (pr_id, name, author_id) 
		VALUES ($1, $2, $3) 
		RETURNING pr_id, name, author_id, status_id
	`
	var p models.PullRequest

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err := conn.QueryRow(ctx, query, pr.PrId, pr.Name, pr.AuthorId).Scan(
		&p.PrId,
		&p.Name,
		&p.AuthorId,
		&p.StatusId,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, repository.ErrAlreadyExists
			}
		}
		return nil, fmt.Errorf("db:PullRequestRepo.Create:QueryRow - %s", err.Error())
	}

	return &p, nil
}

func (r *PullRequestRepo) AddReviewers(ctx context.Context, prId string, reviewersId []string) error {
	query := `
			INSERT INTO pull_requests_reviewers (pr_id, user_id) 
			VALUES ($1, $2)
		`

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	for _, reviewer := range reviewersId {
		_, err := conn.Exec(ctx, query, prId, reviewer)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
                case "23505":
                    return repository.ErrAlreadyExists
                case "23503":
                    return repository.ErrNotFound
                }
			}
			return fmt.Errorf("db:PullRequestRepo.AddReviewers:Exec - %s", err.Error())
		}
	}
	return nil
}

func (r *PullRequestRepo) GetReviewers(ctx context.Context, prId string) ([]string, error) {
	query := `
		SELECT user_id 
		FROM pull_requests_reviewers 
		WHERE pr_id = $1
	`

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, prId)
	if err != nil {
		return nil, fmt.Errorf("db:PullRequestRepo.GetReviewers:Query - %s", err.Error())
	}
	defer rows.Close()

	var reviewersId []string
	for rows.Next() {
		var reviewerId string
		err := rows.Scan(&reviewerId)
		if err != nil {
			return nil, fmt.Errorf("db:PullRequestRepo.GetReviewers:Scan - %s", err.Error())
		}
		reviewersId = append(reviewersId, reviewerId)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db:PullRequestRepo.GetReviewers:rows - %s", err.Error())
	}
	return reviewersId, nil
}

func (r *PullRequestRepo) Merge(ctx context.Context, prId string) (*models.PullRequest, error) {
	query := `
		UPDATE pull_requests 
		SET status_id = 2, merged_at = COALESCE(merged_at, NOW()) 
		WHERE pr_id = $1 
		RETURNING *
	`
	var p models.PullRequest

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err := conn.QueryRow(ctx, query, prId).Scan(
		&p.PrId,
		&p.Name,
		&p.AuthorId,
		&p.StatusId,
		&p.CreatedAt,
		&p.MergedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("db:PullRequestRepo.Merge:QueryRow - %s", err.Error())
	}

	return &p, nil
}

func (r *PullRequestRepo) GetAllReviewByUserId(ctx context.Context, userId string) ([]models.PullRequest, error) {
	query := `
		SELECT pr.pr_id, pr.name, pr.author_id, pr.status_id
		FROM pull_requests as pr 
		JOIN pull_requests_reviewers as r 
		ON pr.pr_id = r.pr_id 
		WHERE r.user_id = $1
	`

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("db:PullRequestRepo.GetAllReviewByUserId:Query - %s", err.Error())
	}
	defer rows.Close()

	var pullRequests []models.PullRequest
	for rows.Next() {
		var pullRequest models.PullRequest
		err := rows.Scan(
			&pullRequest.PrId,
			&pullRequest.Name,
			&pullRequest.AuthorId,
			&pullRequest.StatusId,
		)
		if err != nil {
			return nil, fmt.Errorf("db:PullRequestRepo.GetAllReviewByUserId:Scan - %s", err.Error())
		}
		pullRequests = append(pullRequests, pullRequest)
	}

	return pullRequests, nil
}

func (r *PullRequestRepo) GetStatusById(ctx context.Context, statusId int) (string, error) {
	query := `
		SELECT status 
		FROM pull_requests_statuses 
		WHERE id = $1
	`
	var status string

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err := conn.QueryRow(ctx, query, statusId).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", repository.ErrNotFound
		}
		return "", fmt.Errorf("db:PullRequestRepo.GetStatusById:QueryRow - %s", err.Error())
	}

	return status, nil
}

func (r *PullRequestRepo) GetById(ctx context.Context, prId string) (*models.PullRequest, error) {
	query := `
		SELECT pr_id, name, author_id, status_id
		FROM pull_requests 
		WHERE pr_id = $1
	`
	var pr models.PullRequest

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err := conn.QueryRow(ctx, query, prId).Scan(
		&pr.PrId,
		&pr.Name,
		&pr.AuthorId,
		&pr.StatusId,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("db:PullRequestRepo.GetById:QueryRow - %s", err.Error())
	}

	return &pr, nil
}

func (r *PullRequestRepo) UpdateReviewer(ctx context.Context, prId string, oldReviewer string, newReviewer string) (string, error) {
	query := `
		UPDATE pull_requests_reviewers 
		SET user_id = $1 
		WHERE pr_id = $2 AND user_id = $3 
		RETURNING user_id
	`
	var reviewerId string

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	err := conn.QueryRow(ctx, query, newReviewer, prId, oldReviewer).Scan(&reviewerId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", repository.ErrNotFound
		}
		return "", fmt.Errorf("db:PullRequestRepo.UpdateReviewer:QueryRow - %s", err.Error())
	}

	return reviewerId, nil
}

func (r *PullRequestRepo) GetAllInactiveReviewersByTeam(ctx context.Context, teamId int) ([]models.InactiveReviewers, error) {
	query := `
		SELECT prr.pr_id, prr.user_id 
		FROM pull_requests_reviewers as prr
		JOIN users as u
		ON u.user_id = prr.user_id AND u.is_active = false AND u.team_id = $1
	`

	conn := r.getter.DefaultTrOrDB(ctx, r.db)
	rows, err := conn.Query(ctx, query, teamId)
	if err != nil {
		return nil, fmt.Errorf("db:PullRequestRepo.GetAllInactiveReviewersByTeam:Query - %s", err.Error())
	}
	defer rows.Close()

	var reviewers []models.InactiveReviewers
	for rows.Next() {
		var reviewer models.InactiveReviewers
		err := rows.Scan(
			&reviewer.PrId,
			&reviewer.UserId,
		)
		if err != nil {
			return nil, fmt.Errorf("db:PullRequestRepo.GetAllInactiveReviewersByTeam:Scan - %s", err.Error())
		}
		reviewers = append(reviewers, reviewer)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db:PullRequestRepo.GetAllInactiveReviewersByTeam:rows - %s", err.Error())
	}
	return reviewers, nil
}
