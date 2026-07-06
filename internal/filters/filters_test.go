package filters

import (
	"appointments/internal/assert"
	"testing"
)

func TestLimit(t *testing.T) {
	tests := []struct {
		name    string
		filters Filters
		want    int
	}{
		{name: "page_size_hundred", filters: Filters{PageSize: 100}, want: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.filters.Limit(), tt.want)
		})
	}
}

func TestOffset(t *testing.T) {
	tests := []struct {
		name    string
		filters Filters
		want    int
	}{
		{name: "first_page_zero_offset", filters: Filters{Page: 1, PageSize: 10}, want: 0},
		{name: "second_page", filters: Filters{Page: 2, PageSize: 10}, want: 10},
		{name: "large_page", filters: Filters{Page: 5, PageSize: 25}, want: 100},
		{name: "page_size_one", filters: Filters{Page: 3, PageSize: 1}, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.filters.Offset(), tt.want)
		})
	}
}

func TestSortColumn(t *testing.T) {
	tests := []struct {
		name      string
		filters   Filters
		want      string
		wantPanic bool
	}{
		{name: "valid_sort", filters: Filters{Sort: "title", SortSafeList: buildSafeList()}, want: "title"},
		{name: "desc_sort_trims_minus", filters: Filters{Sort: "-title", SortSafeList: buildSafeList()}, want: "title"},
		{name: "empty_safelist_panics", filters: Filters{Sort: "title"}, wantPanic: true},

		{name: "unsafe_sort_panics", filters: Filters{Sort: "id; DROP TABLE", SortSafeList: buildSafeList()}, wantPanic: true},
		{name: "unsafe_value_empty_safelist_panics", filters: Filters{Sort: "id; DROP TABLE"}, wantPanic: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.wantPanic && r == nil {
					t.Error("expected panic, got none")
				}
				if !tt.wantPanic && r != nil {
					t.Fatalf("unexpected panic: %v", r)
				}
			}()
			assert.Equal(t, tt.filters.SortColumn(), tt.want)
		})
	}
}

func TestSortDirection(t *testing.T) {
	tests := []struct {
		name    string
		filters Filters
		want    string
	}{
		{name: "asc_sort", filters: Filters{Sort: "title"}, want: "ASC"},
		{name: "desc_sort", filters: Filters{Sort: "-title"}, want: "DESC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.filters.SortDirection(), tt.want)
		})
	}
}

func buildSafeList() []string {
	return []string{
		"title", "-title",
		"starts_at", "-starts_at",
		"ends_at", "-ends_at",
		"status", "-status",
	}
}
