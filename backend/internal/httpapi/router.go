package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"onracha/internal/arena"
	"onracha/internal/auth"
	"onracha/internal/game"
)

const sessionCookieName = "onracha_session"

// AuthService describes the authentication operations exposed over HTTP.
type AuthService interface {
	Register(ctx context.Context, name, email, password string) (auth.User, error)
	Login(ctx context.Context, email, password string) (auth.User, auth.Session, error)
	Authenticate(ctx context.Context, token string) (auth.User, error)
}

// ArenaService describes arena operations exposed over HTTP.
type ArenaService interface {
	Create(ctx context.Context, ownerUserID int64, details arena.Details) (arena.Arena, error)
	ListByOwner(ctx context.Context, ownerUserID int64) ([]arena.Arena, error)
	GetByIDAndOwner(ctx context.Context, arenaID, ownerUserID int64) (arena.Arena, error)
	Update(ctx context.Context, arenaID, ownerUserID int64, details arena.Details, active bool) (arena.Arena, error)
	ToggleActive(ctx context.Context, arenaID, ownerUserID int64) (arena.Arena, error)
}

// GameService describes arena availability and game participation operations.
type GameService interface {
	CreateTimeSlot(ctx context.Context, arenaID, ownerUserID int64, startsAt, endsAt time.Time) (game.TimeSlot, error)
	ListAvailableTimeSlots(ctx context.Context, arenaID int64) ([]game.TimeSlot, error)
	Schedule(ctx context.Context, timeSlotID, organizerUserID int64) (game.Game, error)
	ListByArena(ctx context.Context, arenaID int64) ([]game.Game, error)
	Join(ctx context.Context, gameID, userID int64) (game.Game, error)
}

// Config contains browser-facing HTTP settings for the API.
type Config struct {
	AllowedOrigin   string
	SecureCookies   bool
	OpenAPISpecPath string
}

type handler struct {
	authService  AuthService
	arenaService ArenaService
	gameService  GameService
	logger       *slog.Logger
	config       Config
}

// NewRouter builds the API routes and request logging middleware.
func NewRouter(authService AuthService, arenaService ArenaService, gameService GameService, logger *slog.Logger, config Config) http.Handler {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}

	if config.OpenAPISpecPath == "" {
		config.OpenAPISpecPath = "../docs/openapi.yaml"
	}

	api := &handler{authService: authService, arenaService: arenaService, gameService: gameService, logger: logger, config: config}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", api.health)
	mux.HandleFunc("GET /openapi.yaml", api.openAPISpec)
	mux.HandleFunc("GET /swagger", api.swaggerUI)
	mux.HandleFunc("GET /swagger/", api.swaggerUI)
	mux.HandleFunc("POST /api/v1/auth/register", api.register)
	mux.HandleFunc("POST /api/v1/auth/login", api.login)
	mux.HandleFunc("GET /api/v1/auth/me", api.me)
	mux.HandleFunc("GET /api/v1/arenas", api.listArenas)
	mux.HandleFunc("GET /api/v1/arenas/{id}", api.getArena)
	mux.HandleFunc("PUT /api/v1/arenas/{id}", api.updateArena)
	mux.HandleFunc("PATCH /api/v1/arenas/{id}/active", api.toggleArenaActive)
	mux.HandleFunc("POST /api/v1/arenas", api.createArena)
	mux.HandleFunc("POST /api/v1/arenas/{id}/time-slots", api.createTimeSlot)
	mux.HandleFunc("GET /api/v1/arenas/{id}/time-slots", api.listAvailableTimeSlots)
	mux.HandleFunc("GET /api/v1/arenas/{id}/games", api.listGamesByArena)
	mux.HandleFunc("POST /api/v1/time-slots/{id}/games", api.scheduleGame)
	mux.HandleFunc("POST /api/v1/games/{id}/join", api.joinGame)

	return api.cors(api.requestLogger(mux))
}

func (h *handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handler) register(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}

	user, err := h.authService.Register(r.Context(), request.Name, request.Email, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidName), errors.Is(err, auth.ErrInvalidEmail),
			errors.Is(err, auth.ErrPasswordTooShort), errors.Is(err, auth.ErrPasswordTooLong):
			writeError(w, http.StatusBadRequest, "invalid_input", "Confira nome, email e senha.")
		case errors.Is(err, auth.ErrEmailAlreadyRegistered):
			writeError(w, http.StatusConflict, "email_already_registered", "Este email já está cadastrado.")
		default:
			h.logger.Error("account registration failed")
			writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível criar a conta.")
		}
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}

	user, session, err := h.authService.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid_credentials", "Email ou senha inválidos.")
			return
		}
		h.logger.Error("account login failed")
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível entrar.")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		MaxAge:   int(time.Until(session.ExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   h.config.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, user)
}

func (h *handler) me(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authenticatedUser(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *handler) listArenas(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authenticatedUser(w, r)
	if !ok {
		return
	}

	arenas, err := h.arenaService.ListByOwner(r.Context(), user.ID)
	if err != nil {
		h.logger.Error("arena listing failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível listar as arenas.")
		return
	}
	writeJSON(w, http.StatusOK, arenas)
}

func (h *handler) getArena(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authenticatedUser(w, r)
	if !ok {
		return
	}

	arenaID, err := parseArenaID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_arena_id", "O identificador da arena é inválido.")
		return
	}

	value, err := h.arenaService.GetByIDAndOwner(r.Context(), arenaID, user.ID)
	if err != nil {
		if errors.Is(err, arena.ErrNotFound) {
			writeError(w, http.StatusNotFound, "arena_not_found", "Arena não encontrada.")
			return
		}
		h.logger.Error("arena lookup failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível consultar a arena.")
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func (h *handler) updateArena(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authenticatedUser(w, r)
	if !ok {
		return
	}

	arenaID, err := parseArenaID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_arena_id", "O identificador da arena é inválido.")
		return
	}

	var request struct {
		Name      string `json:"name"`
		ArenaType string `json:"arena_type"`
		Address   string `json:"address"`
		City      string `json:"city"`
		State     string `json:"state"`
		Active    *bool  `json:"active"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if request.Active == nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "Informe todos os dados da arena, incluindo o estado de atividade.")
		return
	}

	updatedArena, err := h.arenaService.Update(r.Context(), arenaID, user.ID, arena.Details{
		Name:      request.Name,
		ArenaType: request.ArenaType,
		Address:   request.Address,
		City:      request.City,
		State:     request.State,
	}, *request.Active)
	if err != nil {
		if errors.Is(err, arena.ErrNotFound) {
			writeError(w, http.StatusNotFound, "arena_not_found", "Arena não encontrada.")
			return
		}
		if isInvalidArenaDetails(err) {
			writeError(w, http.StatusBadRequest, "invalid_input", "Confira os dados completos da arena.")
			return
		}
		h.logger.Error("arena update failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível atualizar a arena.")
		return
	}
	writeJSON(w, http.StatusOK, updatedArena)
}

func (h *handler) toggleArenaActive(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authenticatedUser(w, r)
	if !ok {
		return
	}

	arenaID, err := parseArenaID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_arena_id", "O identificador da arena é inválido.")
		return
	}

	updatedArena, err := h.arenaService.ToggleActive(r.Context(), arenaID, user.ID)
	if err != nil {
		if errors.Is(err, arena.ErrNotFound) {
			writeError(w, http.StatusNotFound, "arena_not_found", "Arena não encontrada.")
			return
		}
		h.logger.Error("arena active status toggle failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível alterar o estado da arena.")
		return
	}
	writeJSON(w, http.StatusOK, updatedArena)
}

func (h *handler) createArena(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authenticatedUser(w, r)
	if !ok {
		return
	}

	var request struct {
		Name      string `json:"name"`
		ArenaType string `json:"arena_type"`
		Address   string `json:"address"`
		City      string `json:"city"`
		State     string `json:"state"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}

	createdArena, err := h.arenaService.Create(r.Context(), user.ID, arena.Details{
		Name:      request.Name,
		ArenaType: request.ArenaType,
		Address:   request.Address,
		City:      request.City,
		State:     request.State,
	})
	if err != nil {
		if isInvalidArenaDetails(err) {
			writeError(w, http.StatusBadRequest, "invalid_input", "Confira os dados da arena.")
			return
		}
		h.logger.Error("arena registration failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível cadastrar a arena.")
		return
	}

	writeJSON(w, http.StatusCreated, createdArena)
}

func (h *handler) createTimeSlot(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authenticatedUser(w, r)
	if !ok {
		return
	}

	arenaID, validID := parsePositiveID(r.PathValue("id"))
	if !validID {
		writeError(w, http.StatusBadRequest, "invalid_arena_id", "O identificador da arena é inválido.")
		return
	}

	var request struct {
		StartsAt time.Time `json:"starts_at"`
		EndsAt   time.Time `json:"ends_at"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}

	created, err := h.gameService.CreateTimeSlot(r.Context(), arenaID, user.ID, request.StartsAt, request.EndsAt)
	if err != nil {
		h.writeGameError(w, err, "arena time slot creation failed", "Não foi possível registrar o horário.")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *handler) listAvailableTimeSlots(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.authenticatedUser(w, r); !ok {
		return
	}

	arenaID, validID := parsePositiveID(r.PathValue("id"))
	if !validID {
		writeError(w, http.StatusBadRequest, "invalid_arena_id", "O identificador da arena é inválido.")
		return
	}

	values, err := h.gameService.ListAvailableTimeSlots(r.Context(), arenaID)
	if err != nil {
		h.logger.Error("available time slot listing failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível consultar os horários.")
		return
	}
	writeJSON(w, http.StatusOK, values)
}

func (h *handler) scheduleGame(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authenticatedUser(w, r)
	if !ok {
		return
	}

	timeSlotID, validID := parsePositiveID(r.PathValue("id"))
	if !validID {
		writeError(w, http.StatusBadRequest, "invalid_time_slot_id", "O identificador do horário é inválido.")
		return
	}

	created, err := h.gameService.Schedule(r.Context(), timeSlotID, user.ID)
	if err != nil {
		h.writeGameError(w, err, "game scheduling failed", "Não foi possível agendar o jogo.")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *handler) listGamesByArena(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.authenticatedUser(w, r); !ok {
		return
	}

	arenaID, validID := parsePositiveID(r.PathValue("id"))
	if !validID {
		writeError(w, http.StatusBadRequest, "invalid_arena_id", "O identificador da arena é inválido.")
		return
	}

	values, err := h.gameService.ListByArena(r.Context(), arenaID)
	if err != nil {
		h.logger.Error("game listing failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível consultar os jogos.")
		return
	}
	writeJSON(w, http.StatusOK, values)
}

func (h *handler) joinGame(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authenticatedUser(w, r)
	if !ok {
		return
	}

	gameID, validID := parsePositiveID(r.PathValue("id"))
	if !validID {
		writeError(w, http.StatusBadRequest, "invalid_game_id", "O identificador do jogo é inválido.")
		return
	}

	updated, err := h.gameService.Join(r.Context(), gameID, user.ID)
	if err != nil {
		h.writeGameError(w, err, "game participation failed", "Não foi possível confirmar sua presença.")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *handler) writeGameError(w http.ResponseWriter, err error, logMessage, internalMessage string) {
	switch {
	case errors.Is(err, game.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "Jogo ou horário não encontrado.")
	case errors.Is(err, game.ErrInvalidArenaID), errors.Is(err, game.ErrInvalidTimeSlotID),
		errors.Is(err, game.ErrInvalidGameID), errors.Is(err, game.ErrInvalidUserID),
		errors.Is(err, game.ErrInvalidTimeRange), errors.Is(err, game.ErrTimeSlotInPast):
		writeError(w, http.StatusBadRequest, "invalid_input", "Confira os dados e o horário informado.")
	case errors.Is(err, game.ErrTimeSlotConflict), errors.Is(err, game.ErrTimeSlotUnavailable),
		errors.Is(err, game.ErrUserTimeConflict), errors.Is(err, game.ErrAlreadyParticipating):
		writeError(w, http.StatusConflict, "schedule_conflict", "Esse horário não está disponível para esta operação.")
	default:
		h.logger.Error(logMessage, "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", internalMessage)
	}
}

func parseArenaID(value string) (int64, error) {
	arenaID, valid := parsePositiveID(value)
	if !valid {
		return 0, arena.ErrInvalidArenaID
	}
	return arenaID, nil
}

func parsePositiveID(value string) (int64, bool) {
	id, err := strconv.ParseInt(value, 10, 64)
	return id, err == nil && id > 0
}

func isInvalidArenaDetails(err error) bool {
	return errors.Is(err, arena.ErrInvalidName) ||
		errors.Is(err, arena.ErrInvalidType) ||
		errors.Is(err, arena.ErrInvalidAddress) ||
		errors.Is(err, arena.ErrInvalidCity) ||
		errors.Is(err, arena.ErrInvalidState)
}

func (h *handler) authenticatedUser(w http.ResponseWriter, r *http.Request) (auth.User, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "Entre na sua conta para continuar.")
		return auth.User{}, false
	}

	user, err := h.authService.Authenticate(r.Context(), cookie.Value)
	if errors.Is(err, auth.ErrInvalidSession) {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "Entre na sua conta para continuar.")
		return auth.User{}, false
	}
	if err != nil {
		h.logger.Error("session lookup failed")
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível validar a sessão.")
		return auth.User{}, false
	}
	return user, true
}

func (h *handler) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == h.config.AllowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, OPTIONS")
			w.Header().Add("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *handler) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		response := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(response, r)
		h.logger.Info("http request completed",
			"method", r.Method,
			"route", requestRoute(r),
			"status", response.statusCode,
			"duration_ms", time.Since(startedAt).Milliseconds(),
		)
	})
}

func requestRoute(r *http.Request) string {
	if r.Pattern != "" {
		return r.Pattern
	}

	route := r.Method + " " + r.URL.Path
	switch route {
	case "GET /health", "GET /openapi.yaml", "GET /swagger", "GET /swagger/", "POST /api/v1/auth/register", "POST /api/v1/auth/login", "GET /api/v1/auth/me", "GET /api/v1/arenas", "POST /api/v1/arenas":
		return route
	default:
		return "unmatched"
	}
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusRecorder) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) bool {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Envie os dados como JSON.")
		return false
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "O corpo da requisição é inválido.")
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid_json", "O corpo da requisição é inválido.")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, statusCode int, code, message string) {
	writeJSON(w, statusCode, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
