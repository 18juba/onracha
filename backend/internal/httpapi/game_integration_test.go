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
	"net/url"
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

	arenaRepository := postgres.NewArenaRepository(database)
	arenaService := arena.NewService(arenaRepository)
	cities := []string{"São Paulo", "Campinas"}
	latitudes := []float64{-23.55052, -22.9056}
	longitudes := []float64{-46.633308, -47.0608}
	for index, name := range []string{"First Test Arena", "Second Test Arena"} {
		created, err := arenaService.Create(ctx, users[0].ID, arena.Details{
			Name: name, ArenaType: "society", Address: "Rua do Esporte, 100", City: cities[index], State: "SP",
			Latitude: &latitudes[index], Longitude: &longitudes[index],
		})
		if err != nil {
			t.Fatalf("create arena %d: %v", index, err)
		}
		arenas = append(arenas, created)
	}
	photoURL := "https://images.example.test/sao-paulo.png"
	arenas[0], err = arenaRepository.UpdatePhotoURL(ctx, arenas[0].ID, users[0].ID, photoURL)
	if err != nil {
		t.Fatalf("set arena photo reference for game search: %v", err)
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

	startsAt := time.Now().UTC().Add(72 * time.Hour).Truncate(24 * time.Hour).Add(10 * time.Hour)
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
	nextDaySlot := createSlot(arenas[0].ID, startsAt.Add(24*time.Hour), endsAt.Add(24*time.Hour))

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
	nextDayGameResponse, nextDayGame := schedule(nextDaySlot.ID, secondOrganizerSession)
	if nextDayGameResponse.Code != http.StatusCreated {
		t.Fatalf("schedule next-day game status = %d, want %d; body: %s", nextDayGameResponse.Code, http.StatusCreated, nextDayGameResponse.Body.String())
	}
	freeSlot := createSlot(arenas[0].ID, startsAt.Add(4*time.Hour), endsAt.Add(4*time.Hour))

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
	if len(listedGames) != 3 {
		t.Fatalf("listed games = %d, want 3", len(listedGames))
	}

	myGamesResponse := request(http.MethodGet, "/api/v1/me/games", nil, playerSession)
	if myGamesResponse.Code != http.StatusOK {
		t.Fatalf("list my games status = %d, want %d; body: %s", myGamesResponse.Code, http.StatusOK, myGamesResponse.Body.String())
	}
	var myGames []game.Game
	if err := json.Unmarshal(myGamesResponse.Body.Bytes(), &myGames); err != nil {
		t.Fatalf("decode my games: %v", err)
	}
	if len(myGames) != 2 {
		t.Fatalf("my games = %d, want 2", len(myGames))
	}
	if myGames[0].ID != firstGame.ID || myGames[1].ID != secondGame.ID {
		t.Errorf("my game IDs = [%d, %d], want chronological order [%d, %d]", myGames[0].ID, myGames[1].ID, firstGame.ID, secondGame.ID)
	}
	for _, value := range myGames {
		if value.ArenaID != arenas[0].ID || len(value.Participants) != 2 {
			t.Errorf("my game = %+v, want arena %d and both participants", value, arenas[0].ID)
		}
	}

	type gameSearchResult struct {
		Game  game.Game `json:"game"`
		Arena struct {
			ID        int64    `json:"id"`
			Name      string   `json:"name"`
			Address   string   `json:"address"`
			City      string   `json:"city"`
			State     string   `json:"state"`
			Latitude  *float64 `json:"latitude"`
			Longitude *float64 `json:"longitude"`
			PhotoURL  *string  `json:"photo_url"`
		} `json:"arena"`
	}
	searchByCityAndDatePath := fmt.Sprintf("/api/v1/games?city=%s&date=%s", url.QueryEscape("são paulo"), startsAt.Format("2006-01-02"))
	searchByCityAndDateResponse := request(http.MethodGet, searchByCityAndDatePath, nil, ownerSession)
	if searchByCityAndDateResponse.Code != http.StatusOK {
		t.Fatalf("search games by city and date status = %d, want %d; body: %s", searchByCityAndDateResponse.Code, http.StatusOK, searchByCityAndDateResponse.Body.String())
	}
	var cityAndDateGames []gameSearchResult
	if err := json.Unmarshal(searchByCityAndDateResponse.Body.Bytes(), &cityAndDateGames); err != nil {
		t.Fatalf("decode games filtered by city and date: %v", err)
	}
	if len(cityAndDateGames) != 2 {
		t.Fatalf("games filtered by city and date = %d, want 2", len(cityAndDateGames))
	}
	if cityAndDateGames[0].Game.ID != firstGame.ID || cityAndDateGames[1].Game.ID != secondGame.ID {
		t.Errorf("filtered game IDs = [%d, %d], want [%d, %d]", cityAndDateGames[0].Game.ID, cityAndDateGames[1].Game.ID, firstGame.ID, secondGame.ID)
	}
	for _, value := range cityAndDateGames {
		if value.Arena.City != "São Paulo" || value.Arena.ID != arenas[0].ID || value.Arena.Latitude == nil || *value.Arena.Latitude != latitudes[0] || value.Arena.Longitude == nil || *value.Arena.Longitude != longitudes[0] || value.Arena.PhotoURL == nil || *value.Arena.PhotoURL != photoURL {
			t.Errorf("search result arena = %+v, want arena %d in São Paulo", value.Arena, arenas[0].ID)
		}
	}

	searchByCityResponse := request(http.MethodGet, "/api/v1/games?city=Campinas", nil, ownerSession)
	if searchByCityResponse.Code != http.StatusOK {
		t.Fatalf("search games by city status = %d, want %d; body: %s", searchByCityResponse.Code, http.StatusOK, searchByCityResponse.Body.String())
	}
	var campinasGames []gameSearchResult
	if err := json.Unmarshal(searchByCityResponse.Body.Bytes(), &campinasGames); err != nil {
		t.Fatalf("decode games filtered by Campinas: %v", err)
	}
	if len(campinasGames) != 1 || campinasGames[0].Game.ID != overlappingGame.ID || campinasGames[0].Arena.City != "Campinas" {
		t.Errorf("games filtered by Campinas = %+v, want only game %d in Campinas", campinasGames, overlappingGame.ID)
	}

	searchAvailableForPlayerResponse := request(http.MethodGet, "/api/v1/games", nil, playerSession)
	var availableForPlayer []gameSearchResult
	if searchAvailableForPlayerResponse.Code != http.StatusOK {
		t.Fatalf("search available games for player status = %d, want %d", searchAvailableForPlayerResponse.Code, http.StatusOK)
	}
	if err := json.Unmarshal(searchAvailableForPlayerResponse.Body.Bytes(), &availableForPlayer); err != nil {
		t.Fatalf("decode available games for player: %v", err)
	}
	if len(availableForPlayer) != 1 || availableForPlayer[0].Game.ID != nextDayGame.ID {
		t.Errorf("available games for player = %+v, want only non-overlapping game %d", availableForPlayer, nextDayGame.ID)
	}

	invalidSearchDateResponse := request(http.MethodGet, "/api/v1/games?date=not-a-date", nil, ownerSession)
	if invalidSearchDateResponse.Code != http.StatusBadRequest {
		t.Errorf("invalid search date status = %d, want %d", invalidSearchDateResponse.Code, http.StatusBadRequest)
	}
	unauthenticatedGameSearch := request(http.MethodGet, "/api/v1/games", nil, nil)
	if unauthenticatedGameSearch.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated game search status = %d, want %d", unauthenticatedGameSearch.Code, http.StatusUnauthorized)
	}

	gameDetailsResponse := request(http.MethodGet, fmt.Sprintf("/api/v1/games/%d", firstGame.ID), nil, playerSession)
	if gameDetailsResponse.Code != http.StatusOK {
		t.Fatalf("get game details status = %d, want %d; body: %s", gameDetailsResponse.Code, http.StatusOK, gameDetailsResponse.Body.String())
	}
	var gameDetails game.Game
	if err := json.Unmarshal(gameDetailsResponse.Body.Bytes(), &gameDetails); err != nil {
		t.Fatalf("decode game details: %v", err)
	}
	if gameDetails.ID != firstGame.ID || gameDetails.ArenaID != arenas[0].ID || gameDetails.OrganizerUserID != users[1].ID || len(gameDetails.Participants) != 2 {
		t.Errorf("game details = %+v, want game %d in arena %d with organizer %d and two participants", gameDetails, firstGame.ID, arenas[0].ID, users[1].ID)
	}

	lobbyResponse := request(http.MethodGet, fmt.Sprintf("/api/v1/games/%d/lobby", firstGame.ID), nil, playerSession)
	if lobbyResponse.Code != http.StatusOK {
		t.Fatalf("get game lobby status = %d, want %d; body: %s", lobbyResponse.Code, http.StatusOK, lobbyResponse.Body.String())
	}
	var lobby game.GameLobby
	if err := json.Unmarshal(lobbyResponse.Body.Bytes(), &lobby); err != nil {
		t.Fatalf("decode game lobby: %v", err)
	}
	if lobby.Game.ID != firstGame.ID || lobby.Arena.ID != arenas[0].ID || lobby.Arena.Name != arenas[0].Name || len(lobby.Participants) != 2 {
		t.Fatalf("game lobby = %+v, want game %d, arena %d, and both participants", lobby, firstGame.ID, arenas[0].ID)
	}
	if lobby.Participants[0].Name != users[1].Name || lobby.Participants[1].Name != users[2].Name {
		t.Errorf("lobby participants = %+v, want %q and %q", lobby.Participants, users[1].Name, users[2].Name)
	}

	unauthenticatedGameLobby := request(http.MethodGet, fmt.Sprintf("/api/v1/games/%d/lobby", firstGame.ID), nil, nil)
	if unauthenticatedGameLobby.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated game lobby status = %d, want %d", unauthenticatedGameLobby.Code, http.StatusUnauthorized)
	}

	missingGameResponse := request(http.MethodGet, "/api/v1/games/9223372036854775807", nil, playerSession)
	if missingGameResponse.Code != http.StatusNotFound {
		t.Errorf("missing game details status = %d, want %d", missingGameResponse.Code, http.StatusNotFound)
	}
	invalidGameIDResponse := request(http.MethodGet, "/api/v1/games/not-an-id", nil, playerSession)
	if invalidGameIDResponse.Code != http.StatusBadRequest {
		t.Errorf("invalid game ID status = %d, want %d", invalidGameIDResponse.Code, http.StatusBadRequest)
	}

	unauthenticatedGameDetails := request(http.MethodGet, fmt.Sprintf("/api/v1/games/%d", firstGame.ID), nil, nil)
	if unauthenticatedGameDetails.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated game details status = %d, want %d", unauthenticatedGameDetails.Code, http.StatusUnauthorized)
	}

	unauthenticatedMyGames := request(http.MethodGet, "/api/v1/me/games", nil, nil)
	if unauthenticatedMyGames.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated my games status = %d, want %d", unauthenticatedMyGames.Code, http.StatusUnauthorized)
	}

	leavePath := fmt.Sprintf("/api/v1/games/%d/participants/me", firstGame.ID)
	organizerLeaveResponse := request(http.MethodDelete, leavePath, nil, organizerSession)
	if organizerLeaveResponse.Code != http.StatusConflict {
		t.Errorf("organizer leave status = %d, want %d", organizerLeaveResponse.Code, http.StatusConflict)
	}

	playerLeaveResponse := request(http.MethodDelete, leavePath, nil, playerSession)
	if playerLeaveResponse.Code != http.StatusNoContent {
		t.Fatalf("player leave status = %d, want %d; body: %s", playerLeaveResponse.Code, http.StatusNoContent, playerLeaveResponse.Body.String())
	}
	if playerLeaveResponse.Body.Len() != 0 {
		t.Errorf("player leave body = %q, want empty", playerLeaveResponse.Body.String())
	}

	repeatLeaveResponse := request(http.MethodDelete, leavePath, nil, playerSession)
	if repeatLeaveResponse.Code != http.StatusNoContent {
		t.Errorf("repeat leave status = %d, want %d", repeatLeaveResponse.Code, http.StatusNoContent)
	}

	updatedGameResponse := request(http.MethodGet, fmt.Sprintf("/api/v1/games/%d", firstGame.ID), nil, playerSession)
	var updatedGame game.Game
	if updatedGameResponse.Code != http.StatusOK {
		t.Fatalf("get game after leaving status = %d, want %d", updatedGameResponse.Code, http.StatusOK)
	}
	if err := json.Unmarshal(updatedGameResponse.Body.Bytes(), &updatedGame); err != nil {
		t.Fatalf("decode game after leaving: %v", err)
	}
	if len(updatedGame.Participants) != 1 || updatedGame.Participants[0] != users[1].ID {
		t.Errorf("participants after leaving = %v, want only organizer %d", updatedGame.Participants, users[1].ID)
	}

	updatedMyGamesResponse := request(http.MethodGet, "/api/v1/me/games", nil, playerSession)
	var updatedMyGames []game.Game
	if updatedMyGamesResponse.Code != http.StatusOK {
		t.Fatalf("list my games after leaving status = %d, want %d", updatedMyGamesResponse.Code, http.StatusOK)
	}
	if err := json.Unmarshal(updatedMyGamesResponse.Body.Bytes(), &updatedMyGames); err != nil {
		t.Fatalf("decode my games after leaving: %v", err)
	}
	if len(updatedMyGames) != 1 || updatedMyGames[0].ID != secondGame.ID {
		t.Errorf("my games after leaving = %+v, want only game %d", updatedMyGames, secondGame.ID)
	}

	pastStartsAt := time.Now().UTC().Add(-2 * time.Hour)
	pastEndsAt := time.Now().UTC().Add(-time.Hour)
	if _, err := database.ExecContext(ctx,
		"UPDATE arena_time_slots SET starts_at = $1, ends_at = $2 WHERE id = $3", pastStartsAt, pastEndsAt, secondSlot.ID,
	); err != nil {
		t.Fatalf("mark second game as past: %v", err)
	}
	ongoingStartsAt := time.Now().UTC().Add(-30 * time.Minute)
	ongoingEndsAt := time.Now().UTC().Add(30 * time.Minute)
	if _, err := database.ExecContext(ctx,
		"UPDATE arena_time_slots SET starts_at = $1, ends_at = $2 WHERE id = $3", ongoingStartsAt, ongoingEndsAt, nextDaySlot.ID,
	); err != nil {
		t.Fatalf("mark next-day game as ongoing: %v", err)
	}

	availableGamesDuringOngoing := request(http.MethodGet, "/api/v1/games", nil, playerSession)
	var gamesDuringOngoing []gameSearchResult
	if availableGamesDuringOngoing.Code != http.StatusOK {
		t.Fatalf("search games during an ongoing game status = %d, want %d", availableGamesDuringOngoing.Code, http.StatusOK)
	}
	if err := json.Unmarshal(availableGamesDuringOngoing.Body.Bytes(), &gamesDuringOngoing); err != nil {
		t.Fatalf("decode games during an ongoing game: %v", err)
	}
	foundOngoingGame := false
	for _, result := range gamesDuringOngoing {
		if result.Game.ID == nextDayGame.ID {
			foundOngoingGame = true
		}
		if result.Game.ID == secondGame.ID {
			t.Errorf("past game %d remains in available games", secondGame.ID)
		}
	}
	if !foundOngoingGame {
		t.Errorf("available games during an ongoing game = %+v, want ongoing game %d", gamesDuringOngoing, nextDayGame.ID)
	}

	organizerGamesResponse := request(http.MethodGet, "/api/v1/me/games", nil, organizerSession)
	var organizerGames []game.Game
	if organizerGamesResponse.Code != http.StatusOK {
		t.Fatalf("list organizer games after one ended status = %d, want %d", organizerGamesResponse.Code, http.StatusOK)
	}
	if err := json.Unmarshal(organizerGamesResponse.Body.Bytes(), &organizerGames); err != nil {
		t.Fatalf("decode organizer games: %v", err)
	}
	if len(organizerGames) != 1 || organizerGames[0].ID != firstGame.ID {
		t.Errorf("organizer games after one ended = %+v, want the still-upcoming game %d only", organizerGames, firstGame.ID)
	}

	managedSlotsPath := fmt.Sprintf("/api/v1/arenas/%d/time-slots/manage", arenas[0].ID)
	managedSlotsResponse := request(http.MethodGet, managedSlotsPath, nil, ownerSession)
	var managedSlots []game.ManagedTimeSlot
	if managedSlotsResponse.Code != http.StatusOK {
		t.Fatalf("list managed time slots status = %d, want %d; body: %s", managedSlotsResponse.Code, http.StatusOK, managedSlotsResponse.Body.String())
	}
	if err := json.Unmarshal(managedSlotsResponse.Body.Bytes(), &managedSlots); err != nil {
		t.Fatalf("decode managed time slots: %v", err)
	}
	if len(managedSlots) != 4 {
		t.Errorf("managed time slots = %d, want all 4 slots including free and scheduled slots", len(managedSlots))
	}
	nonOwnerManagedSlots := request(http.MethodGet, managedSlotsPath, nil, organizerSession)
	if nonOwnerManagedSlots.Code != http.StatusNotFound {
		t.Errorf("non-owner managed slot listing status = %d, want %d", nonOwnerManagedSlots.Code, http.StatusNotFound)
	}

	closeFreeSlotPath := fmt.Sprintf("/api/v1/arenas/%d/time-slots/%d/close", arenas[0].ID, freeSlot.ID)
	nonOwnerClose := request(http.MethodPatch, closeFreeSlotPath, nil, organizerSession)
	if nonOwnerClose.Code != http.StatusNotFound {
		t.Errorf("non-owner time slot close status = %d, want %d", nonOwnerClose.Code, http.StatusNotFound)
	}
	closedFreeSlotResponse := request(http.MethodPatch, closeFreeSlotPath, nil, ownerSession)
	var closedFreeSlot game.TimeSlot
	if closedFreeSlotResponse.Code != http.StatusOK {
		t.Fatalf("close available time slot status = %d, want %d; body: %s", closedFreeSlotResponse.Code, http.StatusOK, closedFreeSlotResponse.Body.String())
	}
	if err := json.Unmarshal(closedFreeSlotResponse.Body.Bytes(), &closedFreeSlot); err != nil {
		t.Fatalf("decode closed available time slot: %v", err)
	}
	if closedFreeSlot.ClosedAt == nil {
		t.Fatal("closed available time slot has no closed_at timestamp")
	}
	availableSlotsAfterClose := request(http.MethodGet, fmt.Sprintf("/api/v1/arenas/%d/time-slots", arenas[0].ID), nil, organizerSession)
	var openSlots []game.TimeSlot
	if err := json.Unmarshal(availableSlotsAfterClose.Body.Bytes(), &openSlots); err != nil {
		t.Fatalf("decode available time slots after close: %v", err)
	}
	for _, slot := range openSlots {
		if slot.ID == freeSlot.ID {
			t.Errorf("closed time slot %d remains available for players", freeSlot.ID)
		}
	}
	closedSlotSchedule := request(http.MethodPost, fmt.Sprintf("/api/v1/time-slots/%d/games", freeSlot.ID), nil, organizerSession)
	if closedSlotSchedule.Code != http.StatusConflict {
		t.Errorf("schedule closed time slot status = %d, want %d", closedSlotSchedule.Code, http.StatusConflict)
	}
	closedAgainResponse := request(http.MethodPatch, closeFreeSlotPath, nil, ownerSession)
	if closedAgainResponse.Code != http.StatusOK {
		t.Errorf("closing an already closed time slot status = %d, want %d", closedAgainResponse.Code, http.StatusOK)
	}

	closePastSlotPath := fmt.Sprintf("/api/v1/arenas/%d/time-slots/%d/close", arenas[0].ID, secondSlot.ID)
	closedPastSlot := request(http.MethodPatch, closePastSlotPath, nil, ownerSession)
	if closedPastSlot.Code != http.StatusOK {
		t.Errorf("close past game time slot status = %d, want %d; body: %s", closedPastSlot.Code, http.StatusOK, closedPastSlot.Body.String())
	}

	closeOngoingSlotPath := fmt.Sprintf("/api/v1/arenas/%d/time-slots/%d/close", arenas[0].ID, nextDaySlot.ID)
	closedOngoingSlot := request(http.MethodPatch, closeOngoingSlotPath, nil, ownerSession)
	if closedOngoingSlot.Code != http.StatusOK {
		t.Errorf("close ongoing game time slot status = %d, want %d; body: %s", closedOngoingSlot.Code, http.StatusOK, closedOngoingSlot.Body.String())
	}

	gamesAfterClosing := request(http.MethodGet, "/api/v1/games", nil, playerSession)
	var gamesAfterClosingValues []gameSearchResult
	if err := json.Unmarshal(gamesAfterClosing.Body.Bytes(), &gamesAfterClosingValues); err != nil {
		t.Fatalf("decode games after closing an ongoing game: %v", err)
	}
	for _, result := range gamesAfterClosingValues {
		if result.Game.ID == nextDayGame.ID || result.Game.ID == secondGame.ID {
			t.Errorf("closed or past game %d remains in available game search", result.Game.ID)
		}
	}
	myGamesAfterClosing := request(http.MethodGet, "/api/v1/me/games", nil, secondOrganizerSession)
	var secondOrganizerGames []game.Game
	if err := json.Unmarshal(myGamesAfterClosing.Body.Bytes(), &secondOrganizerGames); err != nil {
		t.Fatalf("decode organizer games after closing an ongoing game: %v", err)
	}
	for _, value := range secondOrganizerGames {
		if value.ID == nextDayGame.ID {
			t.Errorf("closed ongoing game %d remains in the organizer's games", nextDayGame.ID)
		}
	}

	reopenedFreeSlotResponse := request(http.MethodPost, fmt.Sprintf("/api/v1/arenas/%d/time-slots", arenas[0].ID), map[string]time.Time{
		"starts_at": freeSlot.StartsAt, "ends_at": freeSlot.EndsAt,
	}, ownerSession)
	var reopenedFreeSlot game.TimeSlot
	if reopenedFreeSlotResponse.Code != http.StatusCreated {
		t.Fatalf("reopen previously closed time slot status = %d, want %d; body: %s", reopenedFreeSlotResponse.Code, http.StatusCreated, reopenedFreeSlotResponse.Body.String())
	}
	if err := json.Unmarshal(reopenedFreeSlotResponse.Body.Bytes(), &reopenedFreeSlot); err != nil {
		t.Fatalf("decode reopened time slot: %v", err)
	}
	if reopenedFreeSlot.ID != freeSlot.ID || reopenedFreeSlot.ClosedAt != nil {
		t.Errorf("reopened time slot = %+v, want original slot %d with no closed_at", reopenedFreeSlot, freeSlot.ID)
	}

	overlappingOpenSlot := request(http.MethodPost, fmt.Sprintf("/api/v1/arenas/%d/time-slots", arenas[0].ID), map[string]time.Time{
		"starts_at": freeSlot.StartsAt.Add(30 * time.Minute),
		"ends_at":   freeSlot.EndsAt.Add(30 * time.Minute),
	}, ownerSession)
	if overlappingOpenSlot.Code != http.StatusConflict {
		t.Errorf("create overlapping open time slot status = %d, want %d", overlappingOpenSlot.Code, http.StatusConflict)
	}

	closedAgainAfterReopen := request(http.MethodPatch, closeFreeSlotPath, nil, ownerSession)
	if closedAgainAfterReopen.Code != http.StatusOK {
		t.Fatalf("close reopened time slot status = %d, want %d", closedAgainAfterReopen.Code, http.StatusOK)
	}
	overlappingClosedSlot := request(http.MethodPost, fmt.Sprintf("/api/v1/arenas/%d/time-slots", arenas[0].ID), map[string]time.Time{
		"starts_at": freeSlot.StartsAt.Add(30 * time.Minute),
		"ends_at":   freeSlot.EndsAt.Add(30 * time.Minute),
	}, ownerSession)
	if overlappingClosedSlot.Code != http.StatusCreated {
		t.Errorf("create slot overlapping only a closed interval status = %d, want %d; body: %s", overlappingClosedSlot.Code, http.StatusCreated, overlappingClosedSlot.Body.String())
	}

	missingLeaveResponse := request(http.MethodDelete, "/api/v1/games/9223372036854775807/participants/me", nil, playerSession)
	if missingLeaveResponse.Code != http.StatusNotFound {
		t.Errorf("leave missing game status = %d, want %d", missingLeaveResponse.Code, http.StatusNotFound)
	}
	invalidLeaveResponse := request(http.MethodDelete, "/api/v1/games/not-an-id/participants/me", nil, playerSession)
	if invalidLeaveResponse.Code != http.StatusBadRequest {
		t.Errorf("leave with invalid game ID status = %d, want %d", invalidLeaveResponse.Code, http.StatusBadRequest)
	}
	unauthenticatedLeaveResponse := request(http.MethodDelete, leavePath, nil, nil)
	if unauthenticatedLeaveResponse.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated leave status = %d, want %d", unauthenticatedLeaveResponse.Code, http.StatusUnauthorized)
	}
}
