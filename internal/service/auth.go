package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/repository"
	"vitek/internal/tokens"
)

var ErrInvalidMagicLink = errors.New(tokens.ErrMsgInvalidMagicToken)

// MagicLinkMailer delivers the raw magic token (tests use Memory).
type MagicLinkMailer interface {
	SendMagicLink(ctx context.Context, email, rawToken string) error
}

type MemoryMagicLinkMailer struct {
	mu        sync.Mutex
	LastEmail string
	LastToken string
}

func NewMemoryMagicLinkMailer() *MemoryMagicLinkMailer {
	return &MemoryMagicLinkMailer{}
}

func (m *MemoryMagicLinkMailer) SendMagicLink(_ context.Context, email, rawToken string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LastEmail = email
	m.LastToken = rawToken
	return nil
}

type Auth struct {
	pool   *pgxpool.Pool
	mailer MagicLinkMailer
}

func NewAuth(pool *pgxpool.Pool, mailer MagicLinkMailer) *Auth {
	return &Auth{pool: pool, mailer: mailer}
}

func HashOpaqueToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func newOpaqueToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

// RequestMagicLink creates a challenge. ttlOverride < 0 forces already-expired (tests).
func (a *Auth) RequestMagicLink(ctx context.Context, email string, ttlOverride time.Duration) (string, error) {
	q := repository.New(a.pool)
	role := repository.UserRoleUSER
	if u, err := q.GetUserByEmail(ctx, email); err == nil {
		role = u.Role
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	raw, err := newOpaqueToken()
	if err != nil {
		return "", err
	}
	ttl := tokens.MagicLinkTTL
	if ttlOverride != 0 {
		ttl = ttlOverride
	}
	expires := pgtype.Timestamptz{Time: time.Now().UTC().Add(ttl), Valid: true}
	_, err = q.CreateMagicLinkChallenge(ctx, repository.CreateMagicLinkChallengeParams{
		Email:     email,
		TokenHash: HashOpaqueToken(raw),
		RoleHint:  role,
		ExpiresAt: expires,
	})
	if err != nil {
		return "", err
	}
	if err := a.mailer.SendMagicLink(ctx, email, raw); err != nil {
		return "", err
	}
	return raw, nil
}

type SessionUser struct {
	UserID pgtype.UUID
	Email  string
	Role   repository.UserRole
}

func (a *Auth) ConsumeMagicLink(ctx context.Context, rawToken string) (SessionUser, string, error) {
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return SessionUser{}, "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := repository.New(tx)

	chal, err := qtx.ConsumeMagicLinkChallenge(ctx, HashOpaqueToken(rawToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionUser{}, "", ErrInvalidMagicLink
		}
		return SessionUser{}, "", err
	}

	user, err := qtx.GetUserByEmail(ctx, chal.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		user, err = qtx.CreateUserWithRole(ctx, repository.CreateUserWithRoleParams{
			Email: chal.Email,
			Role:  chal.RoleHint,
		})
		if err != nil {
			return SessionUser{}, "", err
		}
		for _, code := range tokens.ShippedServiceCodes() {
			if _, err := qtx.GrantUserService(ctx, repository.GrantUserServiceParams{
				UserID:      user.ID,
				ServiceCode: code,
			}); err != nil {
				return SessionUser{}, "", err
			}
		}
		if _, err := qtx.CreateSubscription(ctx, repository.CreateSubscriptionParams{
			UserID:   user.ID,
			PlanType: repository.PlanTypeFREE,
		}); err != nil {
			return SessionUser{}, "", err
		}
	} else if err != nil {
		return SessionUser{}, "", err
	}

	sessionRaw, err := newOpaqueToken()
	if err != nil {
		return SessionUser{}, "", err
	}
	expires := pgtype.Timestamptz{Time: time.Now().UTC().Add(tokens.SessionTTL), Valid: true}
	_, err = qtx.CreateSession(ctx, repository.CreateSessionParams{
		UserID:    user.ID,
		TokenHash: HashOpaqueToken(sessionRaw),
		ExpiresAt: expires,
	})
	if err != nil {
		return SessionUser{}, "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return SessionUser{}, "", err
	}
	return SessionUser{UserID: user.ID, Email: user.Email, Role: user.Role}, sessionRaw, nil
}

func (a *Auth) SessionFromRaw(ctx context.Context, raw string) (SessionUser, error) {
	if raw == "" {
		return SessionUser{}, ErrInvalidMagicLink
	}
	row, err := repository.New(a.pool).GetActiveSessionByHash(ctx, HashOpaqueToken(raw))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionUser{}, ErrInvalidMagicLink
		}
		return SessionUser{}, err
	}
	return SessionUser{UserID: row.UserID, Email: row.Email, Role: row.Role}, nil
}

func (a *Auth) RevokeSession(ctx context.Context, raw string) error {
	if raw == "" {
		return nil
	}
	return repository.New(a.pool).RevokeSessionByHash(ctx, HashOpaqueToken(raw))
}
