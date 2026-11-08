// Package store is the PostgreSQL persistence layer. Every tenant-scoped query takes clinicID
// explicitly so that cross-tenant access is impossible by construction.
package store

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/almasik/jazyl/internal/domain"
)

// DBTX is satisfied by *pgxpool.Pool and pgx.Tx.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Store wraps a pool or a transaction.
type Store struct {
	db   DBTX
	pool *pgxpool.Pool
}

// New creates a Store on a pool.
func New(pool *pgxpool.Pool) *Store { return &Store{db: pool, pool: pool} }

// Pool exposes the underlying pool (health checks).
func (s *Store) Pool() *pgxpool.Pool { return s.pool }

// InTx runs fn inside a transaction. Nested calls reuse the outer transaction.
func (s *Store) InTx(ctx context.Context, fn func(tx *Store) error) error {
	if _, ok := s.db.(pgx.Tx); ok {
		return fn(s)
	}
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		return fn(&Store{db: tx, pool: s.pool})
	})
}

// mapErr converts pgx errors into domain errors.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("%w: %s", domain.ErrConflict, pgErr.ConstraintName)
		case "23503", "23514", "22P02": // fk, check, invalid text repr
			return fmt.Errorf("%w: %s", domain.ErrInvalid, firstNonEmpty(pgErr.ConstraintName, pgErr.Message))
		}
	}
	return err
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// Page is keyset pagination input: rows strictly older than the cursor.
type Page struct {
	Limit  int
	Cursor string
}

func (p Page) limit() int {
	if p.Limit <= 0 || p.Limit > 200 {
		return 50
	}
	return p.Limit
}

// cursorKey decodes "created_at|id". A zero time means "from the newest".
func (p Page) cursorKey() (time.Time, uuid.UUID, error) {
	if p.Cursor == "" {
		return time.Time{}, uuid.Nil, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(p.Cursor)
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("%w: bad cursor", domain.ErrInvalid)
	}
	ts, id, ok := strings.Cut(string(raw), "|")
	if !ok {
		return time.Time{}, uuid.Nil, fmt.Errorf("%w: bad cursor", domain.ErrInvalid)
	}
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("%w: bad cursor", domain.ErrInvalid)
	}
	u, err := uuid.Parse(id)
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("%w: bad cursor", domain.ErrInvalid)
	}
	return t, u, nil
}

// MakeCursor encodes a keyset cursor.
func MakeCursor(t time.Time, id uuid.UUID) string {
	return base64.RawURLEncoding.EncodeToString([]byte(t.UTC().Format(time.RFC3339Nano) + "|" + id.String()))
}

// keysetArgs returns ($n, $n+1) values; with no cursor uses +infinity so the predicate is a no-op.
func (p Page) keysetArgs() (any, uuid.UUID, error) {
	t, id, err := p.cursorKey()
	if err != nil {
		return nil, uuid.Nil, err
	}
	if t.IsZero() {
		return pgtype.Timestamptz{InfinityModifier: pgtype.Infinity, Valid: true}, uuid.Max, nil
	}
	return t, id, nil
}
