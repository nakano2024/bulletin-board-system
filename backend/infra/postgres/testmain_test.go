package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestMain spins up a disposable Postgres container for the package's infra tests,
// points TEST_DATABASE_URL at it, and tears both down when the tests finish.
func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("bulletin_board"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.WithInitScripts("migrations/0001_create_threads.sql"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to start postgres test container:", err)
		os.Exit(1)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get postgres test container connection string:", err)
		os.Exit(1)
	}

	os.Setenv("TEST_DATABASE_URL", dsn)

	code := m.Run()

	os.Unsetenv("TEST_DATABASE_URL")
	if err := container.Terminate(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "failed to terminate postgres test container:", err)
	}

	os.Exit(code)
}
