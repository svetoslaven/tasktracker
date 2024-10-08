package postgres

import (
	"database/sql"

	"github.com/svetoslaven/tasktracker/internal/repositories"
)

func NewRepositoryRegistry(db *sql.DB) repositories.RepositoryRegistry {
	return repositories.RepositoryRegistry{
		UserRepo:  &UserRepository{DB: db},
		TokenRepo: &TokenRepository{DB: db},
		TeamRepo:  &TeamRepository{DB: db},
		TaskRepo:  &TaskRepository{DB: db},
	}
}
