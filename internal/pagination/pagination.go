package pagination

import (
	"strings"

	"github.com/svetoslaven/tasktracker/internal/validator"
)

const (
	PageKey     = "page"
	PageSizeKey = "page_size"
	SortKey     = "sort"
)

const sortDescendingSuffix = "_desc"

const (
	minPage = 1
	maxPage = 10_000_000

	minPageSize = 1
	maxPageSize = 100
)

type Options struct {
	page         int
	pageSize     int
	sort         string
	sortSafelist []string
}

func NewOptions(page, pageSize int, sort string, sortSafelist []string) Options {
	return Options{
		page:         page,
		pageSize:     pageSize,
		sort:         strings.ToLower(sort),
		sortSafelist: sortSafelist,
	}
}

func (opts Options) Validate(validator *validator.Validator) {
	validator.CheckGreaterThanOrEqualTo(opts.page, minPage, PageKey)
	validator.CheckLessThanOrEqualTo(opts.page, maxPage, PageKey)

	validator.CheckGreaterThanOrEqualTo(opts.pageSize, minPageSize, PageSizeKey)
	validator.CheckLessThanOrEqualTo(opts.pageSize, maxPageSize, PageSizeKey)

	validator.Check(opts.isSortSafe(), SortKey, "Unsupported sort.")
}

func (opts Options) Page() int {
	return opts.page
}

func (opts Options) PageSize() int {
	return opts.pageSize
}

func (opts Options) SortColumn() string {
	if opts.isSortSafe() {
		return strings.TrimSuffix(opts.sort, sortDescendingSuffix)
	}

	panic("unsafe sort parameter: " + opts.sort)
}

func (opts Options) IsSortDescending() bool {
	return strings.HasSuffix(opts.sort, sortDescendingSuffix)
}

func (opts Options) Limit() int {
	return opts.pageSize
}

func (opts Options) Offset() int {
	return (opts.page - 1) * opts.pageSize
}

func (opts Options) isSortSafe() bool {
	for _, safeSort := range opts.sortSafelist {
		safeSort = strings.ToLower(safeSort)

		if opts.sort == safeSort || opts.sort == safeSort+sortDescendingSuffix {
			return true
		}
	}

	return false
}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func CalculateMetadata(page, pageSize, totalRecords int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	lastPage := totalRecords / pageSize

	if totalRecords%pageSize != 0 {
		lastPage++
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     lastPage,
		TotalRecords: totalRecords,
	}
}
