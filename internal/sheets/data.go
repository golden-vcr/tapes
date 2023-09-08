package sheets

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var hexColorRegexp = regexp.MustCompile(`^#[a-zA-Z0-9]{3}(?:[a-zA-Z0-9]{3})?$`)

type RowLookup map[int]Row

type Row struct {
	ID         int
	Title      string
	Year       int
	RuntimeMin int
	Color      string
}

type rowIndexMap struct {
	idColumnIndex      int
	titleColumnIndex   int
	yearColumnIndex    int
	runtimeColumnIndex int
	colorColumnIndex   int
}

func buildRowLookup(result *valuesResult) (RowLookup, error) {
	if len(result.Values) == 0 {
		return nil, fmt.Errorf("sheet data is empty")
	}
	m, err := resolveRowIndices(result.Values[0])
	if err != nil {
		return nil, err
	}

	rows := make(RowLookup)
	for i := 1; i < len(result.Values); i++ {
		values := result.Values[i]
		row, err := parseRow(values, m)
		if err != nil {
			return nil, err
		}
		if _, exists := rows[row.ID]; exists {
			return nil, fmt.Errorf("duplicate ID value: %d", row.ID)
		}
		rows[row.ID] = *row
	}
	return rows, nil
}

func resolveRowIndices(values []string) (rowIndexMap, error) {
	result := rowIndexMap{
		idColumnIndex:      -1,
		titleColumnIndex:   -1,
		yearColumnIndex:    -1,
		runtimeColumnIndex: -1,
		colorColumnIndex:   -1,
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
		if strings.Contains(heading, "id") {
			if err := setIndex("id", &result.idColumnIndex, i); err != nil {
				return result, err
			}
		} else if strings.Contains(heading, "title") {
			if err := setIndex("title", &result.titleColumnIndex, i); err != nil {
				return result, err
			}
		} else if strings.Contains(heading, "year") {
			if err := setIndex("year", &result.yearColumnIndex, i); err != nil {
				return result, err
			}
		} else if strings.Contains(heading, "runtime") {
			if err := setIndex("runtime", &result.runtimeColumnIndex, i); err != nil {
				return result, err
			}
		} else if strings.Contains(heading, "color") {
			if err := setIndex("color", &result.colorColumnIndex, i); err != nil {
				return result, err
			}
		}
	}
	if result.idColumnIndex == -1 {
		return result, fmt.Errorf("could not resolve 'id' column from headings row")
	}
	if result.titleColumnIndex == -1 {
		return result, fmt.Errorf("could not resolve 'title' column from headings row")
	}
	if result.yearColumnIndex == -1 {
		return result, fmt.Errorf("could not resolve 'year' column from headings row")
	}
	if result.runtimeColumnIndex == -1 {
		return result, fmt.Errorf("could not resolve 'runtime' column from headings row")
	}
	if result.colorColumnIndex == -1 {
		return result, fmt.Errorf("could not resolve 'color' column from headings row")
	}
	return result, nil
}

func parseRow(values []string, m rowIndexMap) (*Row, error) {
	idValue := readRowValue(values, m.idColumnIndex)
	if idValue == "" {
		return nil, fmt.Errorf("'id' value is required")
	}
	id, err := strconv.Atoi(idValue)
	if err != nil {
		return nil, fmt.Errorf("'id' value must be an integer (got '%s')", idValue)
	}

	title := readRowValue(values, m.titleColumnIndex)
	if title == "" {
		return nil, fmt.Errorf("'title' value is required")
	}

	year := 0
	yearValue := readRowValue(values, m.yearColumnIndex)
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

	runtime := 0
	runtimeValue := readRowValue(values, m.runtimeColumnIndex)
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

	color := ""
	colorValue := readRowValue(values, m.colorColumnIndex)
	if colorValue != "" {
		if !hexColorRegexp.MatchString(colorValue) {
			return nil, fmt.Errorf("'color' value must be a valid hex color (got '%s')", colorValue)
		}
		color = colorValue
	}

	return &Row{
		ID:         id,
		Title:      title,
		Year:       year,
		RuntimeMin: runtime,
		Color:      color,
	}, nil
}

func readRowValue(values []string, columnIndex int) string {
	if columnIndex >= 0 && columnIndex < len(values) {
		return values[columnIndex]
	}
	return ""
}
