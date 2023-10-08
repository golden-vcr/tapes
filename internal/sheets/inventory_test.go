package sheets

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ListTapes(t *testing.T) {
	tests := []struct {
		name         string
		c            *mockClient
		wantErr      string
		wantWarnings []Warning
		wantTapes    []Tape
	}{
		{
			"API error is a fatal error",
			&mockClient{err: fmt.Errorf("mock error")},
			"failed to get values from inventory spreadsheet: mock error",
			nil,
			nil,
		},
		{
			"normal sheet is parsed OK",
			&mockClient{values: [][]string{
				{"id", "title", "year", "runtime"},
				{"1", "Tape one", "1991", "60"},
				{"2", "Tape two", "", ""},
			}},
			"",
			[]Warning{},
			[]Tape{
				{
					Id:      1,
					Title:   "Tape one",
					Year:    1991,
					Runtime: 60,
					Tags:    []string{},
				},
				{
					Id:    2,
					Title: "Tape two",
					Tags:  []string{},
				},
			},
		},
		{
			"missing columns is a fatal error",
			&mockClient{values: [][]string{
				{"", "title", "year", "runtime"},
				{"1", "Tape one", "1991", "60"},
				{"2", "Tape two", "", ""},
			}},
			"failed to parse headings from first row of inventory spreadsheet: could not resolve 'id' column",
			nil,
			nil,
		},
		{
			"entirely empty spreadsheet is a fatal error",
			&mockClient{values: [][]string{}},
			"inventory spreadsheet has no values",
			nil,
			nil,
		},
		{
			"spreadsheet with valid headings but no rows is OK",
			&mockClient{values: [][]string{
				{"id", "title", "year", "runtime"},
			}},
			"",
			[]Warning{},
			[]Tape{},
		},
		{
			"rows that can't be parsed are ignored and result in a warning",
			&mockClient{values: [][]string{
				{"id", "title", "year", "runtime"},
				{"1", "Tape one", "199X", "60"},
				{"2", "Tape two", "", ""},
			}},
			"",
			[]Warning{
				{
					RowNumber: 2,
					Message:   "'year' value must be an integer (got '199X')",
				},
			},
			[]Tape{
				{
					Id:    2,
					Title: "Tape two",
					Tags:  []string{},
				},
			},
		},
		{
			"tapes with duplicate IDs are ignored and result in a warning",
			&mockClient{values: [][]string{
				{"id", "title", "year", "runtime"},
				{"1", "Tape one", "1991", "60"},
				{"2", "Tape two", "", ""},
				{"2", "Tape three", "", ""},
			}},
			"",
			[]Warning{
				{
					RowNumber: 4,
					Message:   "duplicate tape ID 2: used by both 'Tape three' and 'Tape two'; accepting neither",
				},
			},
			[]Tape{
				{
					Id:      1,
					Title:   "Tape one",
					Year:    1991,
					Runtime: 60,
					Tags:    []string{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tapes, warnings, err := ListTapes(context.Background(), tt.c)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				assert.Nil(t, warnings)
				assert.Nil(t, tapes)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantWarnings, warnings)
				assert.Equal(t, tt.wantTapes, tapes)
			}
		})
	}
}

type mockClient struct {
	err    error
	values [][]string
}

func (m *mockClient) GetValues(ctx context.Context) (*GetValuesResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &GetValuesResult{
		Range:          "ignored",
		MajorDimension: "ignored",
		Values:         m.values,
	}, nil
}

var _ Client = (*mockClient)(nil)
