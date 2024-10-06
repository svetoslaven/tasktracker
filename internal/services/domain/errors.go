package domain

import (
	"errors"

	"github.com/svetoslaven/tasktracker/internal/repositories"
	"github.com/svetoslaven/tasktracker/internal/services"
)

func handleRepositoryRetrievalError(err error) error {
	switch {
	case errors.Is(err, repositories.ErrNoRecordsFound):
		return services.ErrNoRecordsFound
	default:
		return err
	}
}

func handleRepositoryUpdateError(err error) error {
	switch {
	case errors.Is(err, repositories.ErrEditConflict):
		return services.ErrEditConflict
	default:
		return err
	}
}
