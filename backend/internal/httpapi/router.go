package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"onracha/internal/auth"
)

const sessionCookieName = "onracha_session"

// AuthService describes the authentication operations exposed over HTTP.
type AuthService interface {
	Register(ctx context.Context, name, email, password string) (auth.User, error)
	Login(ctx context.Context, email, password string) (auth.User, auth.Session, error)
	Authenticate(ctx context.Context, token string) (auth.User, error)
}

// Config contains browser-facing HTTP settings for the API.
type Config struct {
	AllowedOrigin string
	SecureCookies bool
}

type handler struct {
	authService AuthService
	logger      *slog.Logger
	config      Config
}

// NewRouter builds the API routes and request logging middleware.
func NewRouter(authService AuthService, logger *slog.Logger, config Config) http.Handler {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}

	api := &handler{authService: authService, logger: logger, config: config}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", api.health)
	mux.HandleFunc("POST /api/v1/auth/register", api.register)
	mux.HandleFunc("POST /api/v1/auth/login", api.login)
	mux.HandleFunc("GET /api/v1/auth/me", api.me)

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
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "Entre na sua conta para continuar.")
		return
	}

	user, err := h.authService.Authenticate(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidSession) {
			writeError(w, http.StatusUnauthorized, "unauthenticated", "Entre na sua conta para continuar.")
			return
		}
		h.logger.Error("session lookup failed")
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível validar a sessão.")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *handler) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == h.config.AllowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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
	route := r.Method + " " + r.URL.Path
	switch route {
	case "GET /health", "POST /api/v1/auth/register", "POST /api/v1/auth/login", "GET /api/v1/auth/me":
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
