package postgres

import (
	"context"
	"database/sql"

	"github.com/svetoslaven/tasktracker/internal/repositories"
)

func delete(ctx context.Context, db *sql.DB, query string, args ...any) error {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repositories.ErrNoRecordsFound
	}

	return nil
}
