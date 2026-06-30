package filters

import (
	"appointments/internal/validator"
	"strings"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(checkSortBySafeList(f), "sort", "invalid sort value")
	v.Check(f.PageSize <= 100, "page_size", "must be less than 100 bytes")
	v.Check(f.Page > 0, "page", "must be greater than 0 bytes")
}

func checkSortBySafeList(f Filters) bool {
	for _, safeValue := range f.SortSafeList {
		return safeValue == f.Sort
	}
	return false
}

func (f Filters) Limit() int {
	return f.PageSize
}

func (f Filters) Offset() int {
	return (f.Page - 1) * f.PageSize
}

func (f Filters) SortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	return "created_at"
}

func (f Filters) SortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}
