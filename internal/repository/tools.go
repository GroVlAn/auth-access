package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/GroVlAn/auth-base/ew"
	"github.com/jmoiron/sqlx"
)

func handleQueryError(err error, msg string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ew.New(
			ew.ErrorTypeNotFound,
			err,
		).Msg(msg)
	}

	return ew.New(
		ew.ErrorTypeInternal,
		err,
	)
}

func withTx(ctx context.Context, db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return ew.New(
			ew.ErrorTypeInternal,
			fmt.Errorf("begin tx: %w", err),
		)
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
