package postgres

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	// Load config from .env to get TEST_DB_URL
	viper.SetConfigFile("../../../.env")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading .env file for tests: %s", err)
	}
	testDbUrl := viper.GetString("TEST_DB_URL")
	if testDbUrl == "" {
		log.Fatal("TEST_DB_URL not set in .env file")
	}

	// Connect to the test database
	var err error
	testPool, err = pgxpool.New(context.Background(), testDbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to test database: %v\n", err)
	}
	defer testPool.Close()

	// Run migrations
	mig, err := migrate.New("file://../../../migrations", testDbUrl)
	if err != nil {
		log.Fatalf("Could not create migrate instance: %s", err)
	}
	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Could not run up migrations: %s", err)
	}

	// Run tests
	code := m.Run()

	// Rollback migrations after tests
	if err := mig.Down(); err != nil {
		log.Fatalf("Could not run down migrations: %s", err)
	}

	os.Exit(code)
}
