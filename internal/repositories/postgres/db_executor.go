package postgres

import (
	"context"
	"database/sql"
)

type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
