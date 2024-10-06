package postgres

import "github.com/svetoslaven/tasktracker/internal/pagination"

func CalculateSortDirection(paginationOpts pagination.Options) string {
	if paginationOpts.IsSortDescending() {
		return "DESC"
	} else {
		return "ASC"
	}
}
