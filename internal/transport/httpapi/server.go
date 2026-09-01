package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/starfederation/datastar-go/datastar"

	"vitek/internal/domain"
	"vitek/internal/service"
	"vitek/internal/tokens"
)

type Server struct {
	pool              *pgxpool.Pool
	users             *service.Users
	tasks             *service.Tasks
	proxies           *service.Proxies
	avito             *service.AvitoAccounts
	auth              *service.Auth
	exposeMagicTokens bool
	secureCookies     bool
}

type Option func(*Server)

func WithMagicLinkMailer(m service.MagicLinkMailer) Option {
	return func(s *Server) {
		s.auth = service.NewAuth(s.pool, m)
	}
}

func WithExposeMagicLinkTokens(v bool) Option {
	return func(s *Server) { s.exposeMagicTokens = v }
}

func WithSecureCookies(v bool) Option {
	return func(s *Server) { s.secureCookies = v }
}

func NewServer(pool *pgxpool.Pool, opts ...Option) *Server {
	s := &Server{
		pool:              pool,
		users:             service.NewUsers(pool),
		tasks:             service.NewTasks(pool),
		proxies:           service.NewProxies(pool),
		avito:             service.NewAvitoAccounts(pool),
		auth:              service.NewAuth(pool, service.NewMemoryMagicLinkMailer()),
		exposeMagicTokens: false,
		secureCookies:     false,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(tokens.HTTPGet(tokens.PathHealthz), s.handleHealthz)
	mux.HandleFunc(tokens.HTTPPost(tokens.PathV1Users), s.handleCreateUser)
	mux.HandleFunc(tokens.HTTPPost(tokens.PathV1Tasks), s.handleCreateTask)
	mux.HandleFunc(tokens.HTTPGet(tokens.HTTPPathID(tokens.PathV1Tasks)), s.requireUser(s.handleGetTask))
	mux.HandleFunc(tokens.HTTPGet(tokens.HTTPPathTaskResults()), s.requireUser(s.handleGetTaskResults))
	mux.HandleFunc(tokens.HTTPGet(tokens.PathV1MeTasks), s.requireUser(s.handleListMyTasks))
	mux.HandleFunc(tokens.HTTPPost(tokens.PathV1MeTasks), s.requireUser(s.handleCreateMyTask))
	mux.HandleFunc(tokens.HTTPGet(tokens.PathV1ProxiesActive), s.handleListActiveProxies)
	mux.HandleFunc(tokens.HTTPPost(tokens.PathV1AuthMagicLink), s.handleMagicLinkRequest)
	mux.HandleFunc(tokens.HTTPPost(tokens.PathV1AuthMagicLinkConsume), s.handleMagicLinkConsume)
	mux.HandleFunc(tokens.HTTPGet(tokens.PathV1AuthMagicLinkOpen), s.handleMagicLinkOpen)
	mux.HandleFunc(tokens.HTTPPost(tokens.PathV1AuthLogout), s.handleLogout)
	mux.HandleFunc(tokens.HTTPGet(tokens.PathV1AdminProxies), s.requireAdmin(s.handleAdminListProxies))
	mux.HandleFunc(tokens.HTTPPost(tokens.PathV1AdminProxies), s.requireAdmin(s.handleAdminCreateProxy))
	mux.HandleFunc(tokens.HTTPPatch(tokens.HTTPPathID(tokens.PathV1AdminProxies)), s.requireAdmin(s.handleAdminPatchProxy))
	mux.HandleFunc(tokens.HTTPGet(tokens.PathV1AdminAvitoAccounts), s.requireAdmin(s.handleAdminListAvito))
	mux.HandleFunc(tokens.HTTPPost(tokens.PathV1AdminAvitoAccounts), s.requireAdmin(s.handleAdminCreateAvito))
	mux.HandleFunc(tokens.HTTPPatch(tokens.HTTPPathID(tokens.PathV1AdminAvitoAccounts)), s.requireAdmin(s.handleAdminPatchAvito))
	mux.HandleFunc(tokens.HTTPGet(tokens.PathRoot), s.handleRoot)
	mux.HandleFunc(tokens.HTTPGet(tokens.PathAppSSE), s.requireAdmin(s.handleAppSSE))
	mux.HandleFunc(tokens.HTTPGet(tokens.PathTokensCSS), s.handleDesignCSS)
	return s.withHTTPPolicy(mux)
}

func (s *Server) withHTTPPolicy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !tokens.IsAllowedHTTPRequest(r.Method, r.URL.Path, r.Host) {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := s.sessionUser(r); !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{tokens.JSONFieldError: tokens.ErrMsgUnauthorized})
			return
		}
		next(w, r)
	}
}

func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := s.sessionUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{tokens.JSONFieldError: tokens.ErrMsgUnauthorized})
			return
		}
		if !user.IsAdmin() {
			writeJSON(w, http.StatusForbidden, map[string]string{tokens.JSONFieldError: tokens.ErrMsgForbidden})
			return
		}
		next(w, r)
	}
}

func (s *Server) sessionUser(r *http.Request) (service.SessionUser, bool) {
	c, err := r.Cookie(tokens.CookieSessionName)
	if err != nil || c.Value == "" {
		return service.SessionUser{}, false
	}
	user, err := s.auth.SessionFromRaw(r.Context(), c.Value)
	if err != nil {
		return service.SessionUser{}, false
	}
	return user, true
}

func (s *Server) handleMagicLinkRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidJSON})
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidEmail})
		return
	}
	raw, err := s.auth.RequestMagicLink(r.Context(), req.Email, 0)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{tokens.JSONFieldError: tokens.ErrMsgMagicLinkFailed})
		return
	}
	out := map[string]string{tokens.JSONFieldStatus: tokens.HealthStatusOK}
	if s.exposeMagicTokens {
		out[tokens.JSONFieldToken] = raw
		out[tokens.JSONFieldMagicLinkURL] = tokens.MagicLinkOpenURL(raw)
	}
	writeJSON(w, http.StatusAccepted, out)
}

func (s *Server) handleMagicLinkConsume(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidJSON})
		return
	}
	user, err := s.consumeMagicLink(w, r, strings.TrimSpace(req.Token))
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidMagicToken})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		tokens.JSONFieldEmail: user.Email,
		tokens.JSONFieldRole:  string(user.Role),
		tokens.JSONFieldID:    service.UUIDString(user.UserID),
	})
}

func (s *Server) handleMagicLinkOpen(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get(tokens.QueryParamToken))
	if raw == "" {
		s.writeMagicLinkOpenError(w, http.StatusBadRequest, tokens.MagicLinkOpenCopyMissingToken)
		return
	}
	if _, err := s.consumeMagicLink(w, r, raw); err != nil {
		s.writeMagicLinkOpenError(w, http.StatusUnauthorized, tokens.MagicLinkOpenCopyInvalid)
		return
	}
	http.Redirect(w, r, tokens.PathRoot, http.StatusFound)
}

func (s *Server) consumeMagicLink(w http.ResponseWriter, r *http.Request, rawToken string) (service.SessionUser, error) {
	user, sessionRaw, err := s.auth.ConsumeMagicLink(r.Context(), rawToken)
	if err != nil {
		return service.SessionUser{}, err
	}
	http.SetCookie(w, s.sessionCookie(sessionRaw, time.Now().UTC().Add(tokens.SessionTTL)))
	return user, nil
}

func (s *Server) writeMagicLinkOpenError(w http.ResponseWriter, status int, message string) {
	w.Header().Set(tokens.HeaderContentType, tokens.MIMETextHTML)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(tokens.RenderMagicLinkOpenErrorHTML(message)))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(tokens.CookieSessionName); err == nil && c.Value != "" {
		_ = s.auth.RevokeSession(r.Context(), c.Value)
	}
	clear := s.sessionCookie("", time.Unix(0, 0).UTC())
	clear.MaxAge = -1
	http.SetCookie(w, clear)
	writeJSON(w, http.StatusOK, map[string]string{tokens.JSONFieldStatus: tokens.HealthStatusOK})
}

func (s *Server) sessionCookie(value string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     tokens.CookieSessionName,
		Value:    value,
		Path:     tokens.CookiePath,
		HttpOnly: true,
		Secure:   s.secureCookies,
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
	}
}

func (s *Server) handleAdminListProxies(w http.ResponseWriter, r *http.Request) {
	list, err := s.proxies.ListAll(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{tokens.JSONFieldError: tokens.ErrMsgAdminProxiesFailed})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, p := range list {
		out = append(out, map[string]any{
			tokens.JSONFieldID:       service.UUIDString(p.ID),
			tokens.JSONFieldEndpoint: p.Endpoint,
			tokens.JSONFieldStatus:   string(p.Status),
			tokens.JSONFieldLabel:    p.Label,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{tokens.JSONFieldProxies: out})
}

func (s *Server) handleAdminCreateProxy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Endpoint string              `json:"endpoint"`
		Status   service.ProxyStatus `json:"status"`
		Label    string              `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidJSON})
		return
	}
	if !service.ValidProxyStatus(req.Status) {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidProxyStatus})
		return
	}
	p, err := s.proxies.Create(r.Context(), req.Endpoint, req.Status, req.Label)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{tokens.JSONFieldError: tokens.ErrMsgAdminProxiesFailed})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		tokens.JSONFieldID:       service.UUIDString(p.ID),
		tokens.JSONFieldEndpoint: p.Endpoint,
		tokens.JSONFieldStatus:   string(p.Status),
		tokens.JSONFieldLabel:    p.Label,
	})
}

func (s *Server) handleAdminPatchProxy(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r.PathValue(tokens.PathParamID))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidResourceID})
		return
	}
	var req struct {
		Endpoint string              `json:"endpoint"`
		Status   service.ProxyStatus `json:"status"`
		Label    string              `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidJSON})
		return
	}
	if !service.ValidProxyStatus(req.Status) {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidProxyStatus})
		return
	}
	p, err := s.proxies.Update(r.Context(), id, req.Endpoint, req.Status, req.Label)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{tokens.JSONFieldError: tokens.ErrMsgAdminProxiesFailed})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		tokens.JSONFieldID:       service.UUIDString(p.ID),
		tokens.JSONFieldEndpoint: p.Endpoint,
		tokens.JSONFieldStatus:   string(p.Status),
		tokens.JSONFieldLabel:    p.Label,
	})
}

func (s *Server) handleAdminListAvito(w http.ResponseWriter, r *http.Request) {
	list, err := s.avito.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{tokens.JSONFieldError: tokens.ErrMsgAdminAvitoFailed})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, a := range list {
		out = append(out, map[string]any{
			tokens.JSONFieldID:          service.UUIDString(a.ID),
			tokens.JSONFieldLabel:       a.Label,
			tokens.JSONFieldStatus:      string(a.Status),
			tokens.JSONFieldExternalRef: a.ExternalRef,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{tokens.JSONFieldAccounts: out})
}

func (s *Server) handleAdminCreateAvito(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Label       string                     `json:"label"`
		Status      service.AvitoAccountStatus `json:"status"`
		ExternalRef string                     `json:"external_ref"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidJSON})
		return
	}
	if !service.ValidAvitoStatus(req.Status) {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidAvitoStatus})
		return
	}
	a, err := s.avito.Create(r.Context(), req.Label, req.Status, req.ExternalRef)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{tokens.JSONFieldError: tokens.ErrMsgAdminAvitoFailed})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		tokens.JSONFieldID:          service.UUIDString(a.ID),
		tokens.JSONFieldLabel:       a.Label,
		tokens.JSONFieldStatus:      string(a.Status),
		tokens.JSONFieldExternalRef: a.ExternalRef,
	})
}

func (s *Server) handleAdminPatchAvito(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r.PathValue(tokens.PathParamID))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidResourceID})
		return
	}
	var req struct {
		Label       string                     `json:"label"`
		Status      service.AvitoAccountStatus `json:"status"`
		ExternalRef string                     `json:"external_ref"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidJSON})
		return
	}
	if !service.ValidAvitoStatus(req.Status) {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidAvitoStatus})
		return
	}
	a, err := s.avito.Update(r.Context(), id, req.Label, req.Status, req.ExternalRef)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{tokens.JSONFieldError: tokens.ErrMsgAdminAvitoFailed})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		tokens.JSONFieldID:          service.UUIDString(a.ID),
		tokens.JSONFieldLabel:       a.Label,
		tokens.JSONFieldStatus:      string(a.Status),
		tokens.JSONFieldExternalRef: a.ExternalRef,
	})
}

func (s *Server) handleDesignCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(tokens.HeaderContentType, tokens.MIMETextCSS)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(tokens.RenderDesignCSS()))
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	host := tokens.NormalizeHost(r.Host)
	switch {
	case tokens.IsLandingHost(host):
		s.serveLanding(w, r)
	case tokens.IsAppHost(host):
		s.serveAppPage(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) serveLanding(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(tokens.HeaderContentType, tokens.MIMETextHTML)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(tokens.RenderLandingHTML()))
}

func (s *Server) serveAppPage(w http.ResponseWriter, r *http.Request) {
	html := tokens.RenderAppFaceHTML()
	if user, ok := s.sessionUser(r); ok {
		withSSE := user.IsAdmin()
		html = tokens.RenderAppFaceHTMLLoggedIn(user.Email, service.UUIDString(user.UserID), withSSE)
	}
	w.Header().Set(tokens.HeaderContentType, tokens.MIMETextHTML)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

func (s *Server) handleAppSSE(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	ctx := r.Context()
	ticker := time.NewTicker(tokens.SSETickInterval)
	defer ticker.Stop()

	send := func() error {
		avitoN, err := s.avito.Count(ctx)
		if err != nil {
			return err
		}
		active, err := s.proxies.ListActive(ctx)
		if err != nil {
			return err
		}
		return sse.PatchElements(tokens.AppSSEStatsPatch(
			int(avitoN),
			len(active),
			len(tokens.ShippedServiceCodes()),
		))
	}
	if err := send(); err != nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := send(); err != nil {
				return
			}
		}
	}
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
	Email    string           `json:"email"`
	PlanType service.PlanType `json:"plan_type"`
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
	if !service.ValidPlanType(req.PlanType) {
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
		tokens.JSONFieldID:    service.UUIDString(user.ID),
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
		if errors.Is(err, domain.ErrInvalidListingURL) {
			writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidListingURL})
			return
		}
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
		tokens.JSONFieldID:     service.UUIDString(task.ID),
		tokens.JSONFieldUserID: service.UUIDString(task.UserID),
		tokens.JSONFieldQuery:  task.Query,
		tokens.JSONFieldStatus: string(task.Status),
	})
}

func (s *Server) handleCreateMyTask(w http.ResponseWriter, r *http.Request) {
	user, ok := s.sessionUser(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{tokens.JSONFieldError: tokens.ErrMsgUnauthorized})
		return
	}
	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidJSON})
		return
	}
	task, err := s.tasks.CreateTask(r.Context(), user.UserID, req.Query)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidListingURL) {
			writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidListingURL})
			return
		}
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
	writeJSON(w, http.StatusCreated, service.TaskJSON(task))
}

func (s *Server) handleListMyTasks(w http.ResponseWriter, r *http.Request) {
	user, ok := s.sessionUser(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{tokens.JSONFieldError: tokens.ErrMsgUnauthorized})
		return
	}
	list, err := s.tasks.ListForUser(r.Context(), user.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{tokens.JSONFieldError: tokens.ErrMsgListTasksFailed})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, task := range list {
		out = append(out, service.TaskJSON(task))
	}
	writeJSON(w, http.StatusOK, map[string]any{tokens.JSONFieldTasks: out})
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	user, ok := s.sessionUser(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{tokens.JSONFieldError: tokens.ErrMsgUnauthorized})
		return
	}
	taskID, err := parseUUID(r.PathValue(tokens.PathParamID))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidResourceID})
		return
	}
	task, err := s.tasks.GetForUser(r.Context(), user.UserID, taskID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{tokens.JSONFieldError: tokens.ErrMsgTaskNotFound})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			writeJSON(w, http.StatusForbidden, map[string]string{tokens.JSONFieldError: tokens.ErrMsgForbidden})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{tokens.JSONFieldError: tokens.ErrMsgGetTaskFailed})
		return
	}
	writeJSON(w, http.StatusOK, service.TaskJSON(task))
}

func (s *Server) handleGetTaskResults(w http.ResponseWriter, r *http.Request) {
	user, ok := s.sessionUser(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{tokens.JSONFieldError: tokens.ErrMsgUnauthorized})
		return
	}
	taskID, err := parseUUID(r.PathValue(tokens.PathParamID))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{tokens.JSONFieldError: tokens.ErrMsgInvalidResourceID})
		return
	}
	items, err := s.tasks.ListResultsForUser(r.Context(), user.UserID, taskID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{tokens.JSONFieldError: tokens.ErrMsgTaskNotFound})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			writeJSON(w, http.StatusForbidden, map[string]string{tokens.JSONFieldError: tokens.ErrMsgForbidden})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{tokens.JSONFieldError: tokens.ErrMsgListTasksFailed})
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		out = append(out, service.TaskResultJSON(it))
	}
	writeJSON(w, http.StatusOK, map[string]any{tokens.JSONFieldResults: out})
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
			tokens.JSONFieldID:       service.UUIDString(p.ID),
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
