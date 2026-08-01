package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/GroVlAn/auth-base/ew"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type DBErrorMessages struct {
	Conflict   string
	NotFound   string
	BadRequest string
}

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

func handleDBError(err error, dbErrMSG DBErrorMessages) error {
	var pqErr *pq.Error

	if !errors.As(err, &pqErr) {
		return ew.New(
			ew.ErrorTypeInternal,
			err,
		)
	}

	switch pqErr.Code {
	case "23505":
		return ew.New(
			ew.ErrorTypeConflict,
			err,
		).Msg(dbErrMSG.Conflict)

	case "23503":
		return ew.New(
			ew.ErrorTypeNotFound,
			err,
		).Msg(dbErrMSG.NotFound)

	case "23502", "23514", "22P02":
		if len(dbErrMSG.BadRequest) > 0 {
			return ew.New(
				ew.ErrorTypeBadRequest,
				err,
			).Msg(dbErrMSG.BadRequest)
		}

		return ew.New(
			ew.ErrorTypeInternal,
			err,
		)
	default:
		return ew.New(
			ew.ErrorTypeInternal,
			err,
		)
	}
}
