package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/domain"
	"vitek/internal/repository"
	"vitek/internal/service"
	"vitek/internal/tokens"
)

// Server wires HTTP routes for the control plane.
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
		users:   service.NewUsers(pool),
		tasks:   service.NewTasks(pool),
		proxies: service.NewProxies(q),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(tokens.HTTPGet(tokens.PathHealthz), s.handleHealthz)
	mux.HandleFunc(tokens.HTTPPost(tokens.PathV1Users), s.handleCreateUser)
	mux.HandleFunc(tokens.HTTPPost(tokens.PathV1Tasks), s.handleCreateTask)
	mux.HandleFunc(tokens.HTTPGet(tokens.PathV1ProxiesActive), s.handleListActiveProxies)
	return mux
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if err := s.pool.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			tokens.JSONFieldStatus:  tokens.HealthStatusUnhealthy,
			tokens.JSONFieldProduct: tokens.ProductName,
			tokens.JSONFieldError:   tokens.ErrMsgDatabaseUnavailable,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		tokens.JSONFieldStatus:  tokens.HealthStatusOK,
		tokens.JSONFieldProduct: tokens.ProductName,
	})
}

type createUserRequest struct {
	Email    string              `json:"email"`
	PlanType repository.PlanType `json:"plan_type"`
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidJSON})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidEmail})
		return
	}

	switch req.PlanType {
	case repository.PlanTypeFREE, repository.PlanTypePRO, repository.PlanTypeULTRA:
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidPlanType})
		return
	}

	user, err := s.users.CreateUser(r.Context(), req.Email, req.PlanType)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == tokens.PGCodeUniqueViolation {
			writeJSON(w, http.StatusConflict, map[string]string{tokens.JSONFieldError: tokens.ErrMsgEmailConflict})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{tokens.JSONFieldError: tokens.ErrMsgCreateUserFailed})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		tokens.JSONFieldID:    uuidFromPG(user.ID),
		tokens.JSONFieldEmail: user.Email,
	})
}

type createTaskRequest struct {
	UserID string `json:"user_id"`
	Query  string `json:"query"`
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidJSON})
		return
	}
	userID, err := parseUUID(req.UserID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidUserID})
		return
	}

	task, err := s.tasks.CreateTask(r.Context(), userID, req.Query)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionLimitExceeded) {
			writeJSON(w, http.StatusConflict, map[string]string{tokens.JSONFieldError: domain.ErrSubscriptionLimitExceeded.Error()})
			return
		}
		if errors.Is(err, domain.ErrNoActiveSubscription) {
			writeJSON(w, http.StatusForbidden, map[string]string{tokens.JSONFieldError: domain.ErrNoActiveSubscription.Error()})
			return
		}
		if errors.Is(err, domain.ErrServiceNotEntitled) {
			writeJSON(w, http.StatusForbidden, map[string]string{tokens.JSONFieldError: domain.ErrServiceNotEntitled.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{tokens.JSONFieldError: tokens.ErrMsgCreateTaskFailed})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		tokens.JSONFieldID:     uuidFromPG(task.ID),
		tokens.JSONFieldUserID: uuidFromPG(task.UserID),
		tokens.JSONFieldQuery:  task.Query,
		tokens.JSONFieldStatus: string(task.Status),
	})
}

func (s *Server) handleListActiveProxies(w http.ResponseWriter, r *http.Request) {
	list, err := s.proxies.ListActive(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{tokens.JSONFieldError: tokens.ErrMsgListProxiesFailed})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, p := range list {
		out = append(out, map[string]any{
			tokens.JSONFieldID:       uuidFromPG(p.ID),
			tokens.JSONFieldEndpoint: p.Endpoint,
			tokens.JSONFieldStatus:   string(p.Status),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{tokens.JSONFieldProxies: out})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
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
