package app

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func init() {
	dbURL, ok := os.LookupEnv("DB_URL")
	if !ok {
		panic("app:init:LookupEnv - DB_URL not found")
	}

	dbURL += "?sslmode=disable"
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		panic(fmt.Sprintf("app:init:migrate.New - %s", err.Error()))
	}
	err = m.Up()
	defer m.Close()
	
	if err != nil {
		panic(fmt.Sprintf("app:init:m.Up - %s", err.Error()))
	}
}
