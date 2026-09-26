package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/melojms/mealheiro/internal/store"
)

// coreMaxNameLen caps person, category and tag names (in characters).
const coreMaxNameLen = 40

// coreValidEntryType reports whether t is one of the three entry types.
func coreValidEntryType(t string) bool {
	return t == "expense" || t == "income" || t == "investment"
}

// coreTimestamp returns the current time as the RFC3339 UTC string stored in *_at columns.
func (s *Server) coreTimestamp() string {
	return s.Clock.Now().UTC().Format(time.RFC3339)
}

// coreInTx runs fn inside a transaction, committing on success.
func (s *Server) coreInTx(ctx context.Context, fn func(q *store.Queries) error) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(s.Q.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit()
}

// coreIsUniqueViolation reports whether err is a SQLite UNIQUE constraint failure.
func coreIsUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// coreLikeEscape escapes LIKE wildcards so s matches literally (use with ESCAPE '\').
func coreLikeEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// coreIsNoRows reports whether err is sql.ErrNoRows.
func coreIsNoRows(err error) bool { return errors.Is(err, sql.ErrNoRows) }
