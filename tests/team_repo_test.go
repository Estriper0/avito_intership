package tests

import (
	"github.com/Estriper0/avito_intership/internal/models"
	"github.com/Estriper0/avito_intership/internal/repository"
	"github.com/Estriper0/avito_intership/internal/repository/db"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *TestSuite) TestTeamRepo_Create() {
	repo := db.NewTeamRepo(s.db, trmpgx.DefaultCtxGetter)

	tests := []struct {
		name      string
		teamName  string
		wantID    int
		wantErr   bool
		wantErrIs error
	}{
		{name: "create new team", teamName: "team_1", wantID: 1},
		{name: "create another team", teamName: "team_2", wantID: 2},
		{name: "duplicate team name", teamName: "team_1", wantErr: true, wantErrIs: repository.ErrAlreadyExists},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			id, err := repo.Create(s.ctx, &models.Team{Name: tt.teamName})

			if tt.wantErr {
				require.Error(s.T(), err)
				assert.ErrorIs(s.T(), err, tt.wantErrIs)
				return
			}

			require.NoError(s.T(), err)
			assert.Greater(s.T(), id, 0)
			if tt.wantID > 0 {
				assert.Equal(s.T(), tt.wantID, id)
			}

			//Checking that it was added
			var name string
			err = s.db.QueryRow(s.ctx, "SELECT name FROM teams WHERE id = $1", id).Scan(&name)
			require.NoError(s.T(), err)
			assert.Equal(s.T(), tt.teamName, name)
		})
	}
}

func (s *TestSuite) TestTeamRepo_GetIdByName() {
	repo := db.NewTeamRepo(s.db, trmpgx.DefaultCtxGetter)

	_, err := s.db.Exec(s.ctx, `
		INSERT INTO teams (id, name) VALUES 
			(10, 'team_1'),
			(20, 'team_2'),
			(30, 'team_3')
	`)
	require.NoError(s.T(), err)

	tests := []struct {
		name     string
		teamName string
		wantID   int
		wantErr  bool
	}{
		{name: "existing team", teamName: "team_1", wantID: 10},
		{name: "another existing", teamName: "team_2", wantID: 20},
		{name: "not found", teamName: "unknown-team", wantErr: true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			id, err := repo.GetIdByName(s.ctx, tt.teamName)

			if tt.wantErr {
				require.Error(s.T(), err)
				assert.ErrorIs(s.T(), err, repository.ErrNotFound)
				return
			}

			require.NoError(s.T(), err)
			assert.Equal(s.T(), tt.wantID, id)
		})
	}
}

func (s *TestSuite) TestTeamRepo_GetNameById() {
	repo := db.NewTeamRepo(s.db, trmpgx.DefaultCtxGetter)

	_, err := s.db.Exec(s.ctx, `
		INSERT INTO teams (id, name) VALUES 
			(100, 'team_1'),
			(200, 'team_2')
	`)
	require.NoError(s.T(), err)

	tests := []struct {
		name     string
		teamID   int
		wantName string
		wantErr  bool
	}{
		{name: "existing id", teamID: 100, wantName: "team_1"},
		{name: "another id", teamID: 200, wantName: "team_2"},
		{name: "not found", teamID: 999, wantErr: true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			name, err := repo.GetNameById(s.ctx, tt.teamID)

			if tt.wantErr {
				require.Error(s.T(), err)
				assert.ErrorIs(s.T(), err, repository.ErrNotFound)
				return
			}

			require.NoError(s.T(), err)
			assert.Equal(s.T(), tt.wantName, name)
		})
	}
}
