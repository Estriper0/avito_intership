// pull_request_repo_test.go
package tests

import (
	"fmt"
	"time"

	"github.com/Estriper0/avito_intership/internal/models"
	"github.com/Estriper0/avito_intership/internal/repository"
	"github.com/Estriper0/avito_intership/internal/repository/db"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *TestSuite) TestPullRequestRepo_Create() {
	repo := db.NewPullRequestRepo(s.db, trmpgx.DefaultCtxGetter)

	//Adding a team
	_, err := s.db.Exec(s.ctx, `INSERT INTO teams (id, name) VALUES (1, 'test-team-1')`)
	s.Require().NoError(err)

	//Adding a user
	_, err = s.db.Exec(s.ctx, `INSERT INTO users (user_id, username, team_id, is_active) VALUES ('u1', 'alice', 1, true)`)
	s.Require().NoError(err)

	tests := []struct {
		name    string
		pr      *models.PullRequest
		wantErr bool
		errIs   error
	}{
		{
			name: "create new PR",
			pr:   &models.PullRequest{PrId: "100", Name: "feat/login", AuthorId: "u1"},
		},
		{
			name:    "duplicate pr_id",
			pr:      &models.PullRequest{PrId: "100", Name: "duplicate", AuthorId: "u1"},
			wantErr: true,
			errIs:   repository.ErrAlreadyExists,
		},
		{
			name:    "non-existing author",
			pr:      &models.PullRequest{PrId: "101", Name: "invalid", AuthorId: "ghost"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := repo.Create(s.ctx, tt.pr)

			if tt.wantErr {
				if tt.errIs != nil {
					assert.ErrorIs(s.T(), err, tt.errIs)
				} else {
					require.Error(s.T(), err)
				}
				return
			}

			require.NoError(s.T(), err)
			assert.Equal(s.T(), tt.pr.PrId, got.PrId)
			assert.Equal(s.T(), tt.pr.Name, got.Name)
			assert.Equal(s.T(), tt.pr.AuthorId, got.AuthorId)
			assert.Equal(s.T(), 1, got.StatusId)
		})
	}
}

func (s *TestSuite) TestPullRequestRepo_AddReviewers_GetReviewers() {
	repo := db.NewPullRequestRepo(s.db, trmpgx.DefaultCtxGetter)

	//Adding a team
	_, err := s.db.Exec(s.ctx, `INSERT INTO teams (id, name) VALUES (1, 'test-team-1')`)
	s.Require().NoError(err)

	//Adding a users
	_, err = s.db.Exec(s.ctx, `INSERT INTO users (user_id, username, team_id, is_active) VALUES ('u1', 'alice', 1, true)`)
	s.Require().NoError(err)

	_, err = repo.Create(s.ctx, &models.PullRequest{PrId: "200", Name: "test", AuthorId: "u1"})
	require.NoError(s.T(), err)

	for _, uid := range []string{"rev1", "rev2", "rev3"} {
		//Adding a user
		_, err = s.db.Exec(s.ctx, fmt.Sprintf(`INSERT INTO users (user_id, username, team_id, is_active) VALUES ('%s', '%s', 1, true)`, uid, uid))
		s.Require().NoError(err)
	}

	tests := []struct {
		name        string
		reviewers   []string
		wantErr     bool
		wantErrIs   error
		expectedLen int
	}{
		{name: "add 3 reviewers", reviewers: []string{"rev1", "rev2", "rev3"}, expectedLen: 3},
		{name: "add duplicate", reviewers: []string{"rev2", "rev2"}, wantErr: true, wantErrIs: repository.ErrAlreadyExists},
		{name: "add non-existing user", reviewers: []string{"ghost"}, wantErr: true, wantErrIs: repository.ErrNotFound},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := repo.AddReviewers(s.ctx, "200", tt.reviewers)

			if !tt.wantErr {
				require.NoError(s.T(), err)
			} else {
				require.ErrorIs(s.T(), err, tt.wantErrIs)
			}

			if !tt.wantErr {
				//Cheking GetReviewers
				reviewers, err := repo.GetReviewers(s.ctx, "200")
				require.NoError(s.T(), err)
				assert.ElementsMatch(s.T(), tt.reviewers, reviewers)
			}
		})
	}
}

func (s *TestSuite) TestPullRequestRepo_Merge() {
	repo := db.NewPullRequestRepo(s.db, trmpgx.DefaultCtxGetter)

	//Adding a team
	_, err := s.db.Exec(s.ctx, `INSERT INTO teams (id, name) VALUES (1, 'test-team-1')`)
	s.Require().NoError(err)

	//Adding a users
	_, err = s.db.Exec(s.ctx, `INSERT INTO users (user_id, username, team_id, is_active) VALUES ('u1', 'alice', 1, true)`)
	s.Require().NoError(err)

	_, err = repo.Create(s.ctx, &models.PullRequest{PrId: "300", Name: "to-merge", AuthorId: "u1"})
	require.NoError(s.T(), err)

	tests := []struct {
		name       string
		prId       string
		wantErr    bool
		wantStatus int
	}{
		{name: "merge existing", prId: "300", wantStatus: 2},
		{name: "merge non-existing", prId: "ghost", wantErr: true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			pr, err := repo.Merge(s.ctx, tt.prId)

			if tt.wantErr {
				require.Error(s.T(), err)
				assert.ErrorIs(s.T(), err, repository.ErrNotFound)
				return
			}

			require.NoError(s.T(), err)
			assert.Equal(s.T(), tt.wantStatus, pr.StatusId)
			assert.WithinDuration(s.T(), time.Now(), pr.MergedAt, 5*time.Second)
		})
	}
}

func (s *TestSuite) TestPullRequestRepo_GetAllReviewByUserId() {
	repo := db.NewPullRequestRepo(s.db, trmpgx.DefaultCtxGetter)

	//Adding a team
	_, err := s.db.Exec(s.ctx, `INSERT INTO teams (id, name) VALUES (1, 'test-team-1')`)
	s.Require().NoError(err)

	//Adding a users
	_, err = s.db.Exec(s.ctx, `INSERT INTO users (user_id, username, team_id, is_active) VALUES ('u1', 'alice', 1, true), ('u2', 'reviewer', 1, true)`)
	s.Require().NoError(err)

	// PRs
	for i := 1; i <= 3; i++ {
		_, err := repo.Create(s.ctx, &models.PullRequest{
			PrId:     fmt.Sprintf("pr-%d", i),
			Name:     fmt.Sprintf("pr-%d", i),
			AuthorId: "u1",
		})
		require.NoError(s.T(), err)
	}

	//Adding reviewer
	_, err = s.db.Exec(s.ctx, `
		INSERT INTO pull_requests_reviewers (pr_id, user_id) VALUES
			('pr-1', 'u2'),
			('pr-2', 'u2')
	`)
	require.NoError(s.T(), err)

	prs, err := repo.GetAllReviewByUserId(s.ctx, "u2")
	require.NoError(s.T(), err)
	assert.Len(s.T(), prs, 2)
	assert.Contains(s.T(), []string{"pr-1", "pr-2"}, prs[0].PrId)
	assert.Contains(s.T(), []string{"pr-1", "pr-2"}, prs[1].PrId)
}

func (s *TestSuite) TestPullRequestRepo_GetStatusById() {
	repo := db.NewPullRequestRepo(s.db, trmpgx.DefaultCtxGetter)

	tests := []struct {
		name     string
		statusId int
		want     string
		wantErr  bool
	}{
		{name: "open status", statusId: 1, want: "OPEN"},
		{name: "merged status", statusId: 2, want: "MERGED"},
		{name: "invalid id", statusId: 999, wantErr: true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			status, err := repo.GetStatusById(s.ctx, tt.statusId)

			if tt.wantErr {
				assert.ErrorIs(s.T(), err, repository.ErrNotFound)
				return
			}
			require.NoError(s.T(), err)
			assert.Equal(s.T(), tt.want, status)
		})
	}
}

func (s *TestSuite) TestPullRequestRepo_GetById() {
	repo := db.NewPullRequestRepo(s.db, trmpgx.DefaultCtxGetter)

	//Adding a team
	_, err := s.db.Exec(s.ctx, `INSERT INTO teams (id, name) VALUES (1, 'test-team-1')`)
	s.Require().NoError(err)

	//Adding a users
	_, err = s.db.Exec(s.ctx, `INSERT INTO users (user_id, username, team_id, is_active) VALUES ('u1', 'alice', 1, true)`)
	s.Require().NoError(err)

	_, err = repo.Create(s.ctx, &models.PullRequest{PrId: "500", Name: "find-me", AuthorId: "u1"})
	require.NoError(s.T(), err)

	pr, err := repo.GetById(s.ctx, "500")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "500", pr.PrId)
	assert.Equal(s.T(), "find-me", pr.Name)

	_, err = repo.GetById(s.ctx, "ghost")
	assert.ErrorIs(s.T(), err, repository.ErrNotFound)
}

func (s *TestSuite) TestPullRequestRepo_UpdateReviewer() {
	repo := db.NewPullRequestRepo(s.db, trmpgx.DefaultCtxGetter)

	//Adding a team
	_, err := s.db.Exec(s.ctx, `INSERT INTO teams (id, name) VALUES (1, 'test-team-1')`)
	s.Require().NoError(err)

	//Adding a users
	_, err = s.db.Exec(s.ctx, `INSERT INTO users (user_id, username, team_id, is_active) VALUES ('u1', 'u1', 1, true), ('old-rev', 'old-rev', 1, true), ('new-rev', 'new-rev', 1, true)`)
	s.Require().NoError(err)

	_, err = repo.Create(s.ctx, &models.PullRequest{
		PrId:     "pr-1",
		Name:     "alice",
		AuthorId: "u1",
	})
	require.NoError(s.T(), err)

	_, err = s.db.Exec(s.ctx, `
		INSERT INTO pull_requests_reviewers (pr_id, user_id) VALUES ('pr-1', 'old-rev')
	`)
	require.NoError(s.T(), err)

	newId, err := repo.UpdateReviewer(s.ctx, "pr-1", "old-rev", "new-rev")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "new-rev", newId)

	_, err = repo.UpdateReviewer(s.ctx, "600", "ghost", "someone")
	assert.ErrorIs(s.T(), err, repository.ErrNotFound)
}

func (s *TestSuite) TestPullRequestRepo_GetAllInactiveReviewersByTeam() {
	repo := db.NewPullRequestRepo(s.db, trmpgx.DefaultCtxGetter)

	_, err := s.db.Exec(s.ctx, `
		INSERT INTO teams (id, name) VALUES (1, 'test-team-1');

		INSERT INTO users (user_id, username, team_id, is_active) VALUES
			('u1', 'bob', 1, false),
			('u2', 'charlie', 1, false),
			('u3', 'dave', 1, true);

		INSERT INTO pull_requests (pr_id, name, author_id) VALUES ('pr-1', 'pr-1', 'u3'); 
		
		INSERT INTO pull_requests_reviewers (pr_id, user_id) VALUES
			('pr-1', 'u1'),
			('pr-1', 'u2');
	`)
	require.NoError(s.T(), err)

	reviewers, err := repo.GetAllInactiveReviewersByTeam(s.ctx, 1)
	require.NoError(s.T(), err)

	assert.Len(s.T(), reviewers, 2)	
}
