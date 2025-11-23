package tests

import (
	"fmt"

	"github.com/Estriper0/avito_intership/internal/models"
	"github.com/Estriper0/avito_intership/internal/repository"
	"github.com/Estriper0/avito_intership/internal/repository/db"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *TestSuite) TestUserRepo_CreateOrUpdate() {
	repo := db.NewUserRepo(s.db, trmpgx.DefaultCtxGetter)

	tests := []struct {
		name    string
		user    *models.User
		wantID  string
		wantErr bool
	}{
		{
			name: "create new user",
			user: &models.User{
				UserId:   "user-1",
				Username: "alice",
				TeamId:   10,
				IsActive: true,
			},
			wantID:  "user-1",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			//Adding a team
			_, err := s.db.Exec(s.ctx, fmt.Sprintf(`INSERT INTO teams (id, name) VALUES (%d, 'test-team-1')`, tt.user.TeamId))
			s.Require().NoError(err)

			gotID, err := repo.CreateOrUpdate(s.ctx, tt.user)
			if tt.wantErr {
				s.Require().Error(err)
				return
			}
			require.NoError(s.T(), err)
			assert.Equal(s.T(), tt.wantID, gotID)

			//Checking that it was added
			var u models.User
			err = s.db.QueryRow(s.ctx,
				`SELECT user_id, username, team_id, is_active FROM users WHERE user_id = $1`,
				tt.user.UserId,
			).Scan(&u.UserId, &u.Username, &u.TeamId, &u.IsActive)
			s.Require().NoError(err)
			assert.Equal(s.T(), tt.user.Username, u.Username)
			assert.Equal(s.T(), tt.user.TeamId, u.TeamId)
			assert.Equal(s.T(), tt.user.IsActive, u.IsActive)
		})
	}
}

func (s *TestSuite) TestUserRepo_GetAllByTeam() {
	repo := db.NewUserRepo(s.db, trmpgx.DefaultCtxGetter)

	usersToInsert := []*models.User{
		{UserId: "u1", Username: "u1", TeamId: 5, IsActive: true},
		{UserId: "u2", Username: "u2", TeamId: 5, IsActive: true},
		{UserId: "u3", Username: "u3", TeamId: 6, IsActive: true},
		{UserId: "u4", Username: "u4", TeamId: 5, IsActive: false},
	}

	//Adding a team
	_, err := s.db.Exec(s.ctx, `INSERT INTO teams (id, name) VALUES (5, 'test-team-1')`)
	s.Require().NoError(err)
	_, err = s.db.Exec(s.ctx, `INSERT INTO teams (id, name) VALUES (6, 'test-team-2')`)
	s.Require().NoError(err)

	for _, u := range usersToInsert {
		_, err = repo.CreateOrUpdate(s.ctx, u)
		require.NoError(s.T(), err)
	}

	tests := []struct {
		name    string
		teamID  int
		wantLen int
		wantIDs []string
	}{
		{name: "team 5 - all users", teamID: 5, wantLen: 3, wantIDs: []string{"u1", "u2", "u4"}},
		{name: "team 6", teamID: 6, wantLen: 1, wantIDs: []string{"u3"}},
		{name: "non-existing team", teamID: 999, wantLen: 0, wantIDs: nil},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			users, err := repo.GetAllByTeam(s.ctx, tt.teamID)
			require.NoError(s.T(), err)
			assert.Len(s.T(), users, tt.wantLen)

			if tt.wantLen > 0 {
				gotIDs := make([]string, len(users))
				for i, u := range users {
					gotIDs[i] = u.UserId
				}
				assert.ElementsMatch(s.T(), tt.wantIDs, gotIDs)
			}
		})
	}
}

func (s *TestSuite) TestUserRepo_UpdateIsActive() {
	repo := db.NewUserRepo(s.db, trmpgx.DefaultCtxGetter)

	//Adding a team
	_, err := s.db.Exec(s.ctx, `INSERT INTO teams (id, name) VALUES (1, 'test-team-1')`)
	s.Require().NoError(err)
	_, err = repo.CreateOrUpdate(s.ctx, &models.User{
		UserId: "active-user", Username: "john", TeamId: 1, IsActive: true,
	})
	require.NoError(s.T(), err)

	tests := []struct {
		name        string
		userID      string
		isActive    bool
		wantErr     bool
		wantErrType error
		wantActive  bool
	}{
		{name: "deactivate existing", userID: "active-user", isActive: false, wantActive: false},
		{name: "activate existing", userID: "active-user", isActive: true, wantActive: true},
		{name: "user not found", userID: "ghost", isActive: true, wantErr: true, wantErrType: repository.ErrNotFound},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			user, err := repo.UpdateIsActive(s.ctx, tt.userID, tt.isActive)

			if tt.wantErr {
				require.Error(s.T(), err)
				assert.ErrorIs(s.T(), err, tt.wantErrType)
				return
			}

			require.NoError(s.T(), err)
			assert.NotNil(s.T(), user)
			assert.Equal(s.T(), tt.userID, user.UserId)
			assert.Equal(s.T(), tt.wantActive, user.IsActive)
		})
	}
}

func (s *TestSuite) TestUserRepo_ExistsById() {
	repo := db.NewUserRepo(s.db, trmpgx.DefaultCtxGetter)

	//Adding a team
	_, err := s.db.Exec(s.ctx, `INSERT INTO teams (id, name) VALUES (1, 'test-team-1')`)
	s.Require().NoError(err)

	_, err = repo.CreateOrUpdate(s.ctx, &models.User{UserId: "exists", Username: "test", TeamId: 1, IsActive: true})
	require.NoError(s.T(), err)

	tests := []struct {
		name   string
		userID string
		want   bool
	}{
		{name: "exists", userID: "exists", want: true},
		{name: "not exists", userID: "unknown", want: false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := repo.ExistsById(s.ctx, tt.userID)
			require.NoError(s.T(), err)
			assert.Equal(s.T(), tt.want, got)
		})
	}
}

func (s *TestSuite) TestUserRepo_GetActiveTeamMembersById() {
	repo := db.NewUserRepo(s.db, trmpgx.DefaultCtxGetter)

	//Adding a team
	_, err := s.db.Exec(s.ctx, `INSERT INTO teams (id, name) VALUES (7, 'test-team-7')`)
	s.Require().NoError(err)
	_, err = s.db.Exec(s.ctx, `INSERT INTO teams (id, name) VALUES (8, 'test-team-8')`)
	s.Require().NoError(err)

	fixtures := []*models.User{
		{UserId: "lead", Username: "lead", TeamId: 7, IsActive: true},
		{UserId: "active1", Username: "a1", TeamId: 7, IsActive: true},
		{UserId: "active2", Username: "a2", TeamId: 7, IsActive: true},
		{UserId: "inactive", Username: "in", TeamId: 7, IsActive: false},
		{UserId: "other-team", Username: "ot", TeamId: 8, IsActive: true},
	}
	for _, u := range fixtures {
		_, _ = repo.CreateOrUpdate(s.ctx, u)
	}

	tests := []struct {
		name        string
		userID      string
		wantUserIDs []string
		wantErr     bool
	}{
		{name: "from lead - get two active members", userID: "lead", wantUserIDs: []string{"active1", "active2"}},
		{name: "from active1 - should see lead and active2", userID: "active1", wantUserIDs: []string{"lead", "active2"}},
		{name: "inactive user - still has team, but we return only active", userID: "inactive", wantUserIDs: []string{"lead", "active1", "active2"}},
		{name: "user not found", userID: "ghost", wantErr: false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			users, err := repo.GetActiveTeamMembersById(s.ctx, tt.userID)
			if tt.wantErr {
				require.Error(s.T(), err)
				return
			}
			require.NoError(s.T(), err)

			gotIDs := make([]string, len(users))
			for i, u := range users {
				gotIDs[i] = u.UserId
			}
			assert.ElementsMatch(s.T(), tt.wantUserIDs, gotIDs)
		})
	}
}

func (s *TestSuite) TestUserRepo_MassDeactivation() {
	repo := db.NewUserRepo(s.db, trmpgx.DefaultCtxGetter)

	//Adding a team
	_, err := s.db.Exec(s.ctx, `INSERT INTO teams (id, name) VALUES (1, 'test-team-1')`)
	s.Require().NoError(err)
	existing := []string{"u1", "u2", "u3", "u4"}
	for _, id := range existing {
		_, err = repo.CreateOrUpdate(s.ctx, &models.User{UserId: id, Username: id, TeamId: 1, IsActive: true})
		require.NoError(s.T(), err)
	}

	tests := []struct {
		name        string
		ids         []string
		wantUpdated []string
		wantErr     bool
		wantErrType error
	}{
		{
			name:        "deactivate three existing",
			ids:         []string{"u1", "u2", "u3"},
			wantUpdated: []string{"u1", "u2", "u3"},
		},
		{
			name:        "deactivate some non-existing - only existing ones",
			ids:         []string{"u4", "ghost1", "ghost2"},
			wantUpdated: []string{"u4"},
		},
		{
			name:        "deactivate none - all non-existing",
			ids:         []string{"ghost1", "ghost2"},
			wantErr:     true,
			wantErrType: repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			updatedIDs, err := repo.MassDeactivation(s.ctx, tt.ids)

			if tt.wantErr {
				require.Error(s.T(), err)
				assert.ErrorIs(s.T(), err, tt.wantErrType)
				return
			}

			require.NoError(s.T(), err)
			assert.ElementsMatch(s.T(), tt.wantUpdated, updatedIDs)

			//Check that they are inactive
			for _, id := range updatedIDs {
				var active bool
				err := s.db.QueryRow(s.ctx, `SELECT is_active FROM users WHERE user_id = $1`, id).Scan(&active)
				require.NoError(s.T(), err)
				assert.False(s.T(), active)
			}
		})
	}
}
