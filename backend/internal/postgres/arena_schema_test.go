package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestArenaHasOwnerAndAUserCanOwnMultipleArenas(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set; PostgreSQL integration test is skipped")
	}

	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL connection: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		t.Fatalf("connect to PostgreSQL: %v", err)
	}

	var arenaCount int
	if err := database.QueryRowContext(ctx, "SELECT COUNT(*) FROM arenas").Scan(&arenaCount); err != nil {
		t.Fatalf("query arena table: %v", err)
	}

	email := fmt.Sprintf("arena-schema-%d@example.test", time.Now().UnixNano())
	var userID int64
	err = database.QueryRowContext(ctx,
		`INSERT INTO users (name, email, password_hash) VALUES ('Arena Schema Test', $1, 'test-only-hash') RETURNING id`,
		email,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("create test owner: %v", err)
	}
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := database.ExecContext(cleanupContext, "DELETE FROM arenas WHERE owner_user_id = $1", userID); err != nil {
			t.Errorf("clean up test arenas: %v", err)
		}
		if _, err := database.ExecContext(cleanupContext, "DELETE FROM users WHERE id = $1", userID); err != nil {
			t.Errorf("clean up test owner: %v", err)
		}
	})

	for _, arena := range []struct {
		name     string
		typeName string
	}{
		{name: "Arena Society", typeName: "society"},
		{name: "Arena Futsal", typeName: "futsal"},
	} {
		var arenaID int64
		var createdAt time.Time
		err := database.QueryRowContext(ctx, `
			INSERT INTO arenas (owner_user_id, name, arena_type, address, city, state)
			VALUES ($1, $2, $3, 'Rua do Esporte, 100', 'São Paulo', 'SP')
			RETURNING id, created_at`, userID, arena.name, arena.typeName).Scan(&arenaID, &createdAt)
		if err != nil {
			t.Fatalf("create arena %q: %v", arena.name, err)
		}
		if arenaID == 0 || createdAt.IsZero() {
			t.Errorf("created arena has id %d and created_at %v", arenaID, createdAt)
		}
	}

	if err := database.QueryRowContext(ctx, "SELECT COUNT(*) FROM arenas WHERE owner_user_id = $1", userID).Scan(&arenaCount); err != nil {
		t.Fatalf("count arenas for owner: %v", err)
	}
	if arenaCount != 2 {
		t.Errorf("arenas owned by test user = %d, want 2", arenaCount)
	}

	_, err = database.ExecContext(ctx, `
		INSERT INTO arenas (owner_user_id, name, arena_type, address, city, state)
		VALUES (0, 'Arena sem proprietário', 'society', 'Rua do Esporte, 200', 'São Paulo', 'SP')`)
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "23503" {
		t.Errorf("inserting arena without an existing owner error = %v, want foreign key violation", err)
	}
}
