package sheets

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newIndexMap(t *testing.T) {
	tests := []struct {
		name    string
		values  []string
		wantErr string
		want    indexMap
	}{
		{
			"headings are parsed as expected",
			[]string{"id", "title", "year", "runtime", "contributor"},
			"",
			indexMap{
				idColumnIndex:          0,
				titleColumnIndex:       1,
				yearColumnIndex:        2,
				runtimeColumnIndex:     3,
				contributorColumnIndex: 4,
				columnIndicesByTag:     map[string]int{},
			},
		},
		{
			"tags are parsed as expected",
			[]string{"id", "title", "year", "runtime", "contributor", "Instructional?", "Arts + Crafts?", "History?"},
			"",
			indexMap{
				idColumnIndex:          0,
				titleColumnIndex:       1,
				yearColumnIndex:        2,
				runtimeColumnIndex:     3,
				contributorColumnIndex: 4,
				columnIndicesByTag: map[string]int{
					"instructional": 5,
					"arts+crafts":   6,
					"history":       7,
				},
			},
		},
		{
			"order and extra columns are irrelevant",
			[]string{"", "title", "id", "runtime", "something-else", "padding", "year", "contributor"},
			"",
			indexMap{
				idColumnIndex:          2,
				titleColumnIndex:       1,
				yearColumnIndex:        6,
				runtimeColumnIndex:     3,
				contributorColumnIndex: 7,
				columnIndicesByTag:     map[string]int{},
			},
		},
		{
			"substring match permits mixed case and additional labeling",
			[]string{"ID", " Title: ", "Year (AD)", "Runtime (min.)", "Contributor (Twitch User)"},
			"",
			indexMap{
				idColumnIndex:          0,
				titleColumnIndex:       1,
				yearColumnIndex:        2,
				runtimeColumnIndex:     3,
				contributorColumnIndex: 4,
				columnIndicesByTag:     map[string]int{},
			},
		},
		{
			"all expected columns must be present (even for optional values) or parsing fails",
			[]string{"id", "title", "runtime"},
			"could not resolve 'year' column",
			indexMap{},
		},
		{
			"duplicate column headings will cause parsing to fail",
			[]string{"id", "title", "year", "runtime", "contributor", "Title"},
			"duplicate index for 'title' column",
			indexMap{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newIndexMap(tt.values)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}

}

func Test_rowValues_read(t *testing.T) {
	tests := []struct {
		values      rowValues
		columnIndex int
		want        string
	}{
		{
			[]string{"hello", "world"},
			-1,
			"",
		},
		{
			[]string{"hello", "world"},
			0,
			"hello",
		},
		{
			[]string{"hello", "world"},
			1,
			"world",
		},
		{
			[]string{"hello", "world"},
			2,
			"",
		},
		{
			[]string{"hello", ""},
			1,
			"",
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("in %q, value at %d should be %q", strings.Join(tt.values, ","), tt.columnIndex, tt.want)
		t.Run(name, func(t *testing.T) {
			got := tt.values.read(tt.columnIndex)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseRow(t *testing.T) {
	m := indexMap{0, 1, 2, 3, 4, map[string]int{
		"instructional": 5,
		"history":       6,
	}}
	tests := []struct {
		name    string
		values  rowValues
		wantErr string
		want    *Tape
	}{
		{
			"ordinary tape is parsed OK",
			[]string{"25", "Very cool tape", "1994", "78", "", "1", ""},
			"",
			&Tape{
				Id:      25,
				Title:   "Very cool tape",
				Year:    1994,
				Runtime: 78,
				Tags:    []string{"instructional"},
			},
		},
		{
			"tape with contributor is parsed OK",
			[]string{"25", "Very cool tape", "1994", "78", "12345", "1", ""},
			"",
			&Tape{
				Id:          25,
				Title:       "Very cool tape",
				Year:        1994,
				Runtime:     78,
				Contributor: "12345",
				Tags:        []string{"instructional"},
			},
		},
		{
			"id is required",
			[]string{"", "Very cool tape", "1994", "78", "", "1", ""},
			"'id' value is required",
			nil,
		},
		{
			"id must be an integer",
			[]string{"foo", "Very cool tape", "1994", "78", "", "1", ""},
			"'id' value must be an integer (got 'foo')",
			nil,
		},
		{
			"title is required",
			[]string{"25", "", "1994", "78", "", "1", ""},
			"'title' value is required",
			nil,
		},
		{
			"year is not required and defaults to 0",
			[]string{"25", "Very cool tape", "", "78", "", "1", ""},
			"",
			&Tape{
				Id:      25,
				Title:   "Very cool tape",
				Year:    0,
				Runtime: 78,
				Tags:    []string{"instructional"},
			},
		},
		{
			"year must be an integer if set",
			[]string{"25", "Very cool tape", "1988.5", "78", "", "1", ""},
			"'year' value must be an integer (got '1988.5')",
			nil,
		},
		{
			"runtime is not required and defaults to 0",
			[]string{"25", "Very cool tape", "1994", "", "", "", "1"},
			"",
			&Tape{
				Id:      25,
				Title:   "Very cool tape",
				Year:    1994,
				Runtime: 0,
				Tags:    []string{"history"},
			},
		},
		{
			"runtime must be an integer if set",
			[]string{"25", "Very cool tape", "1994", "4h", "", "1", ""},
			"'runtime' value must be an integer (got '4h')",
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.parseRow(tt.values)
			if tt.wantErr != "" {
				assert.Nil(t, got)
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
