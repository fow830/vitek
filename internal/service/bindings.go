package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/domain"
	"vitek/internal/repository"
	"vitek/internal/tokens"
)

// Bindings links Avito accounts to sticky proxies + Rod profiles.
type Bindings struct {
	pool *pgxpool.Pool
	q    *repository.Queries
}

func NewBindings(pool *pgxpool.Pool) *Bindings {
	return &Bindings{pool: pool, q: repository.New(pool)}
}

func (s *Bindings) Create(ctx context.Context, accountID, proxyID pgtype.UUID, userDataDir string) (repository.ListingFetchBinding, error) {
	userDataDir = strings.TrimSpace(userDataDir)
	if userDataDir == "" {
		userDataDir = tokens.FixtureBindingUserDataDir
	}
	return s.q.CreateBinding(ctx, repository.CreateBindingParams{
		AvitoAccountID: accountID,
		ProxyID:        proxyID,
		UserDataDir:    userDataDir,
	})
}

func (s *Bindings) List(ctx context.Context) ([]repository.ListingFetchBinding, error) {
	return s.q.ListBindings(ctx)
}

func (s *Bindings) UpdateStatus(ctx context.Context, id pgtype.UUID, statusRaw string) (repository.ListingFetchBinding, error) {
	var status repository.ListingBindingStatus
	switch statusRaw {
	case tokens.ListingBindingStatusActive:
		status = repository.ListingBindingStatusACTIVE
	case tokens.ListingBindingStatusPaused:
		status = repository.ListingBindingStatusPAUSED
	case tokens.ListingBindingStatusDisabled:
		status = repository.ListingBindingStatusDISABLED
	default:
		return repository.ListingFetchBinding{}, fmt.Errorf("%s", tokens.ErrMsgInvalidBindingStatus)
	}
	return s.q.UpdateBindingStatus(ctx, repository.UpdateBindingStatusParams{ID: id, Status: status})
}

func (s *Bindings) Disable(ctx context.Context, id pgtype.UUID) error {
	_, err := s.UpdateStatus(ctx, id, tokens.ListingBindingStatusDisabled)
	return err
}

func (s *Bindings) PauseForProxy(ctx context.Context, proxyID pgtype.UUID) error {
	return s.q.PauseBindingsForProxy(ctx, proxyID)
}

// MarkSessionReady forces READY (tests / explicit admin bypass).
func (s *Bindings) MarkSessionReady(ctx context.Context, id pgtype.UUID) (repository.ListingFetchBinding, error) {
	return s.q.UpdateBindingSession(ctx, repository.UpdateBindingSessionParams{
		ID:            id,
		SessionStatus: repository.ListingSessionStatusREADY,
		SessionErr:    nil,
	})
}

func (s *Bindings) MarkSessionChallenge(ctx context.Context, id pgtype.UUID, errMsg string) (repository.ListingFetchBinding, error) {
	msg := strings.TrimSpace(errMsg)
	if msg == "" {
		msg = tokens.ErrMsgSessionChallenge
	}
	return s.q.UpdateBindingSession(ctx, repository.UpdateBindingSessionParams{
		ID:            id,
		SessionStatus: repository.ListingSessionStatusCHALLENGE,
		SessionErr:    &msg,
	})
}

// WarmSession opens Avito via the binding profile+proxy and marks READY / CHALLENGE / ERROR.
func (s *Bindings) WarmSession(ctx context.Context, id pgtype.UUID, fetch AvitoPageFetcher) (repository.ListingFetchBinding, error) {
	if fetch == nil {
		return repository.ListingFetchBinding{}, fmt.Errorf("%s", tokens.ErrMsgSessionWarmFailed)
	}
	b, err := s.q.GetBinding(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return repository.ListingFetchBinding{}, domain.ErrResourceNotFound
		}
		return repository.ListingFetchBinding{}, err
	}
	proxy, err := s.q.GetProxy(ctx, b.ProxyID)
	if err != nil {
		return repository.ListingFetchBinding{}, err
	}

	if _, err := s.q.UpdateBindingSession(ctx, repository.UpdateBindingSessionParams{
		ID:            id,
		SessionStatus: repository.ListingSessionStatusLOGGINGIN,
		SessionErr:    nil,
	}); err != nil {
		return repository.ListingFetchBinding{}, err
	}

	html, err := fetch.FetchHTML(ctx, proxy.Endpoint, b.UserDataDir, tokens.ProxyProbeURL())
	if err != nil {
		msg := err.Error()
		failed, updErr := s.q.UpdateBindingSession(ctx, repository.UpdateBindingSessionParams{
			ID:            id,
			SessionStatus: repository.ListingSessionStatusERROR,
			SessionErr:    &msg,
		})
		if updErr != nil {
			return repository.ListingFetchBinding{}, updErr
		}
		return failed, fmt.Errorf("%s: %w", tokens.ErrMsgSessionWarmFailed, err)
	}
	if tokens.IsAvitoChallengeHTML(html) {
		msg := tokens.ErrMsgSessionChallenge
		return s.q.UpdateBindingSession(ctx, repository.UpdateBindingSessionParams{
			ID:            id,
			SessionStatus: repository.ListingSessionStatusCHALLENGE,
			SessionErr:    &msg,
		})
	}
	return s.q.UpdateBindingSession(ctx, repository.UpdateBindingSessionParams{
		ID:            id,
		SessionStatus: repository.ListingSessionStatusREADY,
		SessionErr:    nil,
	})
}

func (s *Bindings) PickReady(ctx context.Context) (repository.PickReadyBindingRow, error) {
	row, err := s.q.PickReadyBinding(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			return repository.PickReadyBindingRow{}, domain.ErrListingSearchNoBinding
		}
		return repository.PickReadyBindingRow{}, err
	}
	return row, nil
}

func BindingJSON(b repository.ListingFetchBinding) map[string]any {
	out := map[string]any{
		tokens.JSONFieldID:            UUIDString(b.ID),
		tokens.JSONFieldAccountID:     UUIDString(b.AvitoAccountID),
		tokens.JSONFieldProxyID:       UUIDString(b.ProxyID),
		tokens.JSONFieldUserDataDir:   b.UserDataDir,
		tokens.JSONFieldStatus:        string(b.Status),
		tokens.JSONFieldSessionStatus: string(b.SessionStatus),
	}
	if b.SessionErr != nil {
		out[tokens.JSONFieldSessionErr] = *b.SessionErr
	}
	return out
}

func ProxyJSON(p repository.Proxy) map[string]any {
	out := map[string]any{
		tokens.JSONFieldID:         UUIDString(p.ID),
		tokens.JSONFieldEndpoint:   p.Endpoint,
		tokens.JSONFieldStatus:     string(p.Status),
		tokens.JSONFieldLabel:      p.Label,
		tokens.JSONFieldHealth:     string(p.Health),
		tokens.JSONFieldFailStreak: p.FailStreak,
	}
	if p.LastOkAt.Valid {
		out[tokens.JSONFieldLastOkAt] = p.LastOkAt.Time
	}
	if p.LastErr != nil {
		out[tokens.JSONFieldLastErr] = *p.LastErr
	}
	return out
}
