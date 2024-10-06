package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/svetoslaven/tasktracker/internal/repositories"
)

func isDuplicateKeyError(err error, constraint string) bool {
	return err.Error() == fmt.Sprintf(`pq: duplicate key value violates unique constraint "%s"`, constraint)
}

func handleQueryRowError(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return repositories.ErrNoRecordsFound
	default:
		return err
	}
}
