package postgres

import (
	"context"
	"database/sql"
)

func runInTransaction(ctx context.Context, db *sql.DB, txOpts *sql.TxOptions, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	err = tx.Commit()
	return err
}
