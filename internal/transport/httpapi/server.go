package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/domain"
	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
)

// Server wires HTTP routes for Phase B control plane.
type Server struct {
	pool    *pgxpool.Pool
	users   *service.Users
	tasks   *service.Tasks
	proxies *service.Proxies
}

func NewServer(pool *pgxpool.Pool) *Server {
	q := repository.New(pool)
	return &Server{
		pool:    pool,
		users:   service.NewUsers(q),
		tasks:   service.NewTasks(pool),
		proxies: service.NewProxies(q),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("POST /v1/users", s.handleCreateUser)
	mux.HandleFunc("POST /v1/tasks", s.handleCreateTask)
	mux.HandleFunc("GET /v1/proxies/active", s.handleListActiveProxies)
	return mux
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if err := s.pool.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status":  "unhealthy",
			"product": tokens.ProductName,
			"error":   "database unavailable",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"product": tokens.ProductName,
	})
}

type createUserRequest struct {
	Email string `json:"email"`
	Plan  string `json:"plan"`
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	plan := repository.PlanType(req.Plan)
	switch plan {
	case repository.PlanTypeFREE, repository.PlanTypePRO, repository.PlanTypeULTRA:
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid plan"})
		return
	}

	user, err := s.users.CreateUser(r.Context(), req.Email, plan)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create user failed"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":    uuidFromPG(user.ID),
		"email": user.Email,
	})
}

type createTaskRequest struct {
	UserID string `json:"user_id"`
	Query  string `json:"query"`
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	userID, err := parseUUID(req.UserID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user_id"})
		return
	}

	task, err := s.tasks.CreateTask(r.Context(), userID, req.Query)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionLimitExceeded) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": domain.ErrSubscriptionLimitExceeded.Error()})
			return
		}
		if errors.Is(err, domain.ErrNoActiveSubscription) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": domain.ErrNoActiveSubscription.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create task failed"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":      uuidFromPG(task.ID),
		"user_id": uuidFromPG(task.UserID),
		"query":   task.Query,
		"status":  string(task.Status),
	})
}

func (s *Server) handleListActiveProxies(w http.ResponseWriter, r *http.Request) {
	list, err := s.proxies.ListActive(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list proxies failed"})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, p := range list {
		out = append(out, map[string]any{
			"id":       uuidFromPG(p.ID),
			"endpoint": p.Endpoint,
			"status":   string(p.Status),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"proxies": out})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseUUID(raw string) (pgtype.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

func uuidFromPG(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return uuid.UUID(id.Bytes).String()
}

// Ping is used by tests/helpers.
func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}
