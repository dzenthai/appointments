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
