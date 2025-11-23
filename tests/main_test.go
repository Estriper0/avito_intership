package tests

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	pg "github.com/Estriper0/avito_intership/pkg/postgres"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestSuite struct {
	suite.Suite

	ctx         context.Context
	db          *pgxpool.Pool
	pgContainer *postgres.PostgresContainer
}

func TestRepositoriesSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupSuite() {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18.1-alpine3.22",
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	s.Require().NoError(err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	s.Require().NoError(err)

	db, err := pg.New(connStr, 10)
	s.Require().NoError(err)

	migrations(connStr)

	s.ctx = ctx
	s.db = db
	s.pgContainer = pgContainer
}

func (s *TestSuite) TearDownSuite() {
	s.db.Close()
	s.pgContainer.Terminate(s.ctx)
}

func (s *TestSuite) SetupTest() {
	_, err := s.db.Exec(s.ctx, "TRUNCATE TABLE teams, users, pull_requests, pull_requests_reviewers CASCADE;")
	s.Require().NoError(err)
}

func migrations(dbUrl string) {
	m, err := migrate.New("file://../migrations", dbUrl)
	if err != nil {
		panic(fmt.Sprintf("test:migrations:migrate.New - %s", err.Error()))
	}
	err = m.Up()
	defer m.Close()

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(fmt.Sprintf("test:migrations:m.Up - %s", err.Error()))
	}
}
