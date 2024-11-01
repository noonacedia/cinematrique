package data

import (
	"github.com/noonacedia/cinematrique/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be positive")
	v.Check(f.Page <= 10_000_000, "page", "must be less then 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be positive")
	v.Check(f.PageSize <= 100, "page_size", "must be less then 100")
	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}
