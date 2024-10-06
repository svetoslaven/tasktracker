package postgres

import (
	"fmt"
)

func isDuplicateKeyError(err error, constraint string) bool {
	return err.Error() == fmt.Sprintf(`pq: duplicate key value violates unique constraint "%s"`, constraint)
}
