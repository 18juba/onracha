package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"onracha/internal/arena"
	"onracha/internal/auth"
	"onracha/internal/game"
	"onracha/internal/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestGameSchedulingEnforcesArenaAndUserTimeRules(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set; PostgreSQL integration test is skipped")
	}

	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL connection: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		t.Fatalf("connect to PostgreSQL: %v", err)
	}

	authService := auth.NewService(postgres.NewAuthRepository(database), auth.NewArgon2idHasher())
	users := make([]auth.User, 0, 4)
	sessions := make([]auth.Session, 0, 4)
	arenas := make([]arena.Arena, 0, 2)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		for _, value := range arenas {
			if _, err := database.ExecContext(cleanupContext, `
				DELETE FROM games
				WHERE time_slot_id IN (SELECT id FROM arena_time_slots WHERE arena_id = $1)`, value.ID); err != nil {
				t.Errorf("clean up games for arena %d: %v", value.ID, err)
			}
			if _, err := database.ExecContext(cleanupContext, "DELETE FROM arena_time_slots WHERE arena_id = $1", value.ID); err != nil {
				t.Errorf("clean up time slots for arena %d: %v", value.ID, err)
			}
			if _, err := database.ExecContext(cleanupContext, "DELETE FROM arenas WHERE id = $1", value.ID); err != nil {
				t.Errorf("clean up arena %d: %v", value.ID, err)
			}
		}
		for _, user := range users {
			if _, err := database.ExecContext(cleanupContext, "DELETE FROM users WHERE id = $1", user.ID); err != nil {
				t.Errorf("clean up user %d: %v", user.ID, err)
			}
		}
	})
	for index, name := range []string{"Arena Owner", "First Organizer", "Joining Player", "Second Organizer"} {
		email := fmt.Sprintf("game-flow-%d-%d@example.test", index, time.Now().UnixNano())
		user, err := authService.Register(ctx, name, email, "Senha123")
		if err != nil {
			t.Fatalf("register %s: %v", name, err)
		}
		users = append(users, user)
		_, session, err := authService.Login(ctx, email, "Senha123")
		if err != nil {
			t.Fatalf("login %s: %v", name, err)
		}
		sessions = append(sessions, session)
	}

	arenaService := arena.NewService(postgres.NewArenaRepository(database))
	for index, name := range []string{"First Test Arena", "Second Test Arena"} {
		created, err := arenaService.Create(ctx, users[0].ID, arena.Details{
			Name: name, ArenaType: "society", Address: "Rua do Esporte, 100", City: "São Paulo", State: "SP",
		})
		if err != nil {
			t.Fatalf("create arena %d: %v", index, err)
		}
		arenas = append(arenas, created)
	}

	gameService := game.NewService(postgres.NewTimeSlotRepository(database), postgres.NewGameRepository(database))
	api := NewRouter(authService, arenaService, gameService, slog.New(slog.NewTextHandler(io.Discard, nil)), Config{})
	ownerSession := &sessions[0]
	organizerSession := &sessions[1]
	playerSession := &sessions[2]
	secondOrganizerSession := &sessions[3]
	request := func(method, path string, body any, session *auth.Session) *httptest.ResponseRecorder {
		t.Helper()
		var requestBody io.Reader
		if body != nil {
			encoded, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("encode request body: %v", err)
			}
			requestBody = bytes.NewReader(encoded)
		}
		httpRequest := httptest.NewRequest(method, path, requestBody)
		if body != nil {
			httpRequest.Header.Set("Content-Type", "application/json")
		}
		if session != nil {
			httpRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: session.Token})
		}
		response := httptest.NewRecorder()
		api.ServeHTTP(response, httpRequest)
		return response
	}

	startsAt := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Minute)
	endsAt := startsAt.Add(time.Hour)
	createSlot := func(arenaID int64, start, end time.Time) game.TimeSlot {
		t.Helper()
		response := request(http.MethodPost, fmt.Sprintf("/api/v1/arenas/%d/time-slots", arenaID), map[string]time.Time{
			"starts_at": start, "ends_at": end,
		}, ownerSession)
		if response.Code != http.StatusCreated {
			t.Fatalf("create time slot status = %d, want %d; body: %s", response.Code, http.StatusCreated, response.Body.String())
		}
		var value game.TimeSlot
		if err := json.Unmarshal(response.Body.Bytes(), &value); err != nil {
			t.Fatalf("decode created time slot: %v", err)
		}
		return value
	}

	firstSlot := createSlot(arenas[0].ID, startsAt, endsAt)
	secondSlot := createSlot(arenas[0].ID, startsAt.Add(2*time.Hour), endsAt.Add(2*time.Hour))
	overlappingSlot := createSlot(arenas[1].ID, startsAt.Add(30*time.Minute), endsAt.Add(30*time.Minute))

	unauthorizedSlot := request(http.MethodPost, fmt.Sprintf("/api/v1/arenas/%d/time-slots", arenas[0].ID), map[string]time.Time{
		"starts_at": startsAt.Add(4 * time.Hour), "ends_at": endsAt.Add(4 * time.Hour),
	}, organizerSession)
	if unauthorizedSlot.Code != http.StatusNotFound {
		t.Fatalf("non-owner time slot creation status = %d, want %d", unauthorizedSlot.Code, http.StatusNotFound)
	}

	conflictingArenaSlot := request(http.MethodPost, fmt.Sprintf("/api/v1/arenas/%d/time-slots", arenas[0].ID), map[string]time.Time{
		"starts_at": startsAt.Add(30 * time.Minute), "ends_at": endsAt.Add(30 * time.Minute),
	}, ownerSession)
	if conflictingArenaSlot.Code != http.StatusConflict {
		t.Fatalf("overlapping arena time slot status = %d, want %d", conflictingArenaSlot.Code, http.StatusConflict)
	}

	availableResponse := request(http.MethodGet, fmt.Sprintf("/api/v1/arenas/%d/time-slots", arenas[0].ID), nil, organizerSession)
	if availableResponse.Code != http.StatusOK {
		t.Fatalf("list available slots status = %d, want %d", availableResponse.Code, http.StatusOK)
	}
	var availableSlots []game.TimeSlot
	if err := json.Unmarshal(availableResponse.Body.Bytes(), &availableSlots); err != nil {
		t.Fatalf("decode available slots: %v", err)
	}
	if len(availableSlots) != 2 {
		t.Fatalf("available slots = %d, want 2", len(availableSlots))
	}

	schedule := func(slotID int64, session *auth.Session) (*httptest.ResponseRecorder, game.Game) {
		t.Helper()
		response := request(http.MethodPost, fmt.Sprintf("/api/v1/time-slots/%d/games", slotID), nil, session)
		var value game.Game
		if response.Code == http.StatusCreated {
			if err := json.Unmarshal(response.Body.Bytes(), &value); err != nil {
				t.Fatalf("decode scheduled game: %v", err)
			}
		}
		return response, value
	}

	firstGameResponse, firstGame := schedule(firstSlot.ID, organizerSession)
	if firstGameResponse.Code != http.StatusCreated {
		t.Fatalf("schedule first game status = %d, want %d; body: %s", firstGameResponse.Code, http.StatusCreated, firstGameResponse.Body.String())
	}
	if firstGame.OrganizerUserID != users[1].ID || len(firstGame.Participants) != 1 || firstGame.Participants[0] != users[1].ID {
		t.Fatalf("first game = %+v, want organizer automatically enrolled", firstGame)
	}

	duplicateSlotResponse := request(http.MethodPost, fmt.Sprintf("/api/v1/time-slots/%d/games", firstSlot.ID), nil, secondOrganizerSession)
	if duplicateSlotResponse.Code != http.StatusConflict {
		t.Fatalf("second game for slot status = %d, want %d", duplicateSlotResponse.Code, http.StatusConflict)
	}

	organizerTimeConflict, _ := schedule(overlappingSlot.ID, organizerSession)
	if organizerTimeConflict.Code != http.StatusConflict {
		t.Fatalf("organizer overlapping game status = %d, want %d", organizerTimeConflict.Code, http.StatusConflict)
	}

	secondGameResponse, secondGame := schedule(secondSlot.ID, organizerSession)
	if secondGameResponse.Code != http.StatusCreated {
		t.Fatalf("schedule game at different time status = %d, want %d; body: %s", secondGameResponse.Code, http.StatusCreated, secondGameResponse.Body.String())
	}

	overlappingGameResponse, overlappingGame := schedule(overlappingSlot.ID, secondOrganizerSession)
	if overlappingGameResponse.Code != http.StatusCreated {
		t.Fatalf("schedule overlapping slot by another user status = %d, want %d; body: %s", overlappingGameResponse.Code, http.StatusCreated, overlappingGameResponse.Body.String())
	}

	join := func(gameID int64, session *auth.Session) (*httptest.ResponseRecorder, game.Game) {
		t.Helper()
		response := request(http.MethodPost, fmt.Sprintf("/api/v1/games/%d/join", gameID), nil, session)
		var value game.Game
		if response.Code == http.StatusOK {
			if err := json.Unmarshal(response.Body.Bytes(), &value); err != nil {
				t.Fatalf("decode joined game: %v", err)
			}
		}
		return response, value
	}

	firstJoinResponse, firstJoinedGame := join(firstGame.ID, playerSession)
	if firstJoinResponse.Code != http.StatusOK {
		t.Fatalf("join first game status = %d, want %d; body: %s", firstJoinResponse.Code, http.StatusOK, firstJoinResponse.Body.String())
	}
	if len(firstJoinedGame.Participants) != 2 {
		t.Fatalf("participants after first join = %d, want 2", len(firstJoinedGame.Participants))
	}

	duplicateJoin, _ := join(firstGame.ID, playerSession)
	if duplicateJoin.Code != http.StatusConflict {
		t.Fatalf("duplicate join status = %d, want %d", duplicateJoin.Code, http.StatusConflict)
	}

	playerTimeConflict, _ := join(overlappingGame.ID, playerSession)
	if playerTimeConflict.Code != http.StatusConflict {
		t.Fatalf("overlapping game join status = %d, want %d", playerTimeConflict.Code, http.StatusConflict)
	}

	secondJoinResponse, secondJoinedGame := join(secondGame.ID, playerSession)
	if secondJoinResponse.Code != http.StatusOK {
		t.Fatalf("join game at different time status = %d, want %d; body: %s", secondJoinResponse.Code, http.StatusOK, secondJoinResponse.Body.String())
	}
	if len(secondJoinedGame.Participants) != 2 {
		t.Fatalf("participants at different time = %d, want 2", len(secondJoinedGame.Participants))
	}

	gamesResponse := request(http.MethodGet, fmt.Sprintf("/api/v1/arenas/%d/games", arenas[0].ID), nil, playerSession)
	if gamesResponse.Code != http.StatusOK {
		t.Fatalf("list games status = %d, want %d", gamesResponse.Code, http.StatusOK)
	}
	var listedGames []game.Game
	if err := json.Unmarshal(gamesResponse.Body.Bytes(), &listedGames); err != nil {
		t.Fatalf("decode listed games: %v", err)
	}
	if len(listedGames) != 2 {
		t.Fatalf("listed games = %d, want 2", len(listedGames))
	}
}
