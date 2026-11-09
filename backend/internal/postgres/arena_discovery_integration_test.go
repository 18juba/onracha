package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"onracha/internal/arena"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestArenaRepositoryDiscoverIncludesArenasWithoutGames(t *testing.T) {
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

	email := fmt.Sprintf("arena-discover-%d@example.test", time.Now().UnixNano())
	var ownerUserID int64
	if err := database.QueryRowContext(ctx,
		`INSERT INTO users (name, email, password_hash) VALUES ('Arena Discovery Test', $1, 'test-only-hash') RETURNING id`,
		email,
	).Scan(&ownerUserID); err != nil {
		t.Fatalf("create test owner: %v", err)
	}
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := database.ExecContext(cleanupContext, "DELETE FROM arenas WHERE owner_user_id = $1", ownerUserID); err != nil {
			t.Errorf("clean up test arenas: %v", err)
		}
		if _, err := database.ExecContext(cleanupContext, "DELETE FROM users WHERE id = $1", ownerUserID); err != nil {
			t.Errorf("clean up test owner: %v", err)
		}
	})

	repository := NewArenaRepository(database)
	testCity := fmt.Sprintf("Arena Discovery %d", time.Now().UnixNano())
	nearLatitude, nearLongitude := -23.551, -46.631
	farLatitude, farLongitude := -24.0, -46.63
	createArena := func(name string, latitude, longitude *float64) arena.Arena {
		t.Helper()
		created, err := repository.Create(ctx, arena.Arena{
			OwnerUserID: ownerUserID,
			Name:        name,
			ArenaType:   "society",
			Address:     "Rua do Esporte, 100",
			City:        testCity,
			State:       "SP",
			Latitude:    latitude,
			Longitude:   longitude,
			Active:      true,
		})
		if err != nil {
			t.Fatalf("create test arena %q: %v", name, err)
		}
		return created
	}

	nearArena := createArena("Nearby Arena Without Games", &nearLatitude, &nearLongitude)
	farArena := createArena("Far Arena Without Games", &farLatitude, &farLongitude)
	noLocationArena := createArena("Arena Without Coordinates", nil, nil)
	inactiveArena := createArena("Inactive Arena", &nearLatitude, &nearLongitude)
	if _, err := repository.ToggleActive(ctx, inactiveArena.ID, ownerUserID); err != nil {
		t.Fatalf("deactivate test arena: %v", err)
	}

	cityResults, err := repository.Discover(ctx, arena.DiscoveryFilters{City: strings.ToLower(testCity)})
	if err != nil {
		t.Fatalf("discover arenas by city: %v", err)
	}
	if len(cityResults) != 3 {
		t.Fatalf("city discovery returned %d arenas, want 3 active arenas without regard to scheduled games", len(cityResults))
	}
	cityResultIDs := map[int64]bool{}
	for _, value := range cityResults {
		cityResultIDs[value.ID] = true
	}
	if !cityResultIDs[nearArena.ID] || !cityResultIDs[farArena.ID] || !cityResultIDs[noLocationArena.ID] || cityResultIDs[inactiveArena.ID] {
		t.Errorf("city discovery arena IDs = %v; want active city arenas, including one without coordinates, and no inactive arena", cityResultIDs)
	}

	radiusKM := 10.0
	nearbyResults, err := repository.Discover(ctx, arena.DiscoveryFilters{
		City: testCity, Latitude: &nearLatitude, Longitude: &nearLongitude, RadiusKM: &radiusKM,
	})
	if err != nil {
		t.Fatalf("discover arenas near location: %v", err)
	}
	if len(nearbyResults) != 1 || nearbyResults[0].ID != nearArena.ID {
		t.Errorf("nearby discovery = %+v, want only arena %d", nearbyResults, nearArena.ID)
	}
}
