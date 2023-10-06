package sheets

import (
	"fmt"
	"strconv"
	"strings"
)

// In the Golden VCR Inventory spreadsheet, the columns that we want to parse are
// labeled in the first row with these values, which we check for using a
// case-insensitive substring search
const (
	columnHeadingSubstringId      = "id"
	columnHeadingSubstringTitle   = "title"
	columnHeadingSubstringYear    = "year"
	columnHeadingSubstringRuntime = "runtime"
)

// indexMap is a lookup that tells us which column index in the spreadsheet (0 for A,
// 1 for B, etc.) contains the each of the values we want to parse
type indexMap struct {
	idColumnIndex      int
	titleColumnIndex   int
	yearColumnIndex    int
	runtimeColumnIndex int
}

// newIndexMap builds an indexMap given the values in the first row of a spreadsheet, or
// returns an error if unable to find a matching column heading for each value
func newIndexMap(values []string) (indexMap, error) {
	m := indexMap{
		idColumnIndex:      -1,
		titleColumnIndex:   -1,
		yearColumnIndex:    -1,
		runtimeColumnIndex: -1,
	}
	setIndex := func(name string, p *int, value int) error {
		if *p >= 0 {
			return fmt.Errorf("duplicate index for '%s' column", name)
		}
		*p = value
		return nil
	}
	for i, value := range values {
		heading := strings.ToLower(value)
		if strings.Contains(heading, columnHeadingSubstringId) {
			if err := setIndex("id", &m.idColumnIndex, i); err != nil {
				return m, err
			}
		} else if strings.Contains(heading, columnHeadingSubstringTitle) {
			if err := setIndex("title", &m.titleColumnIndex, i); err != nil {
				return m, err
			}
		} else if strings.Contains(heading, columnHeadingSubstringYear) {
			if err := setIndex("year", &m.yearColumnIndex, i); err != nil {
				return m, err
			}
		} else if strings.Contains(heading, columnHeadingSubstringRuntime) {
			if err := setIndex("runtime", &m.runtimeColumnIndex, i); err != nil {
				return m, err
			}
		}
	}
	if m.idColumnIndex == -1 {
		return m, fmt.Errorf("could not resolve 'id' column")
	}
	if m.titleColumnIndex == -1 {
		return m, fmt.Errorf("could not resolve 'title' column")
	}
	if m.yearColumnIndex == -1 {
		return m, fmt.Errorf("could not resolve 'year' column")
	}
	if m.runtimeColumnIndex == -1 {
		return m, fmt.Errorf("could not resolve 'runtime' column")
	}
	return m, nil
}

// rowValues is an array of values representing a single row in a spreadsheet, with the
// string at index 0 representing the value in column A, the string at 1 representing B,
// and so on
type rowValues []string

// read resolves the value contained in the row at the given column, defaulting to an
// empty string if the column index is out of bounds
func (v rowValues) read(columnIndex int) string {
	if columnIndex >= 0 && columnIndex < len(v) {
		return v[columnIndex]
	}
	return ""
}

// parseRow attempts to resolve a valid Tape struct from a row in the spreadsheet,
// returning an error if the row could not be parsed due to unexpected format, missing
// data in required columns, etc.
func (m *indexMap) parseRow(values rowValues) (*Tape, error) {
	// Integer 'id' is required: note that we don't check for uniqueness here
	idValue := values.read(m.idColumnIndex)
	if idValue == "" {
		return nil, fmt.Errorf("'id' value is required")
	}
	id, err := strconv.Atoi(idValue)
	if err != nil {
		return nil, fmt.Errorf("'id' value must be an integer (got '%s')", idValue)
	}

	// String 'title' is required
	title := values.read(m.titleColumnIndex)
	if title == "" {
		return nil, fmt.Errorf("'title' value is required")
	}

	// Integer 'year' is optional; default to 0 if not set
	year := 0
	yearValue := values.read(m.yearColumnIndex)
	if yearValue != "" {
		yearAsInt, err := strconv.Atoi(yearValue)
		if err != nil {
			return nil, fmt.Errorf("'year' value must be an integer (got '%s')", yearValue)
		}
		if yearAsInt <= 0 {
			return nil, fmt.Errorf("'year' int value must be positive (got %d)", yearAsInt)
		}
		year = yearAsInt
	}

	// Integer 'runtime' is optional; default to 0 if not set
	runtime := 0
	runtimeValue := values.read(m.runtimeColumnIndex)
	if runtimeValue != "" {
		runtimeAsInt, err := strconv.Atoi(runtimeValue)
		if err != nil {
			return nil, fmt.Errorf("'runtime' value must be an integer (got '%s')", runtimeValue)
		}
		if runtimeAsInt <= 0 {
			return nil, fmt.Errorf("'runtime' int value must be positive (got %d)", runtimeAsInt)
		}
		runtime = runtimeAsInt
	}

	return &Tape{
		Id:      id,
		Title:   title,
		Year:    year,
		Runtime: runtime,
	}, nil
}
