package sheets

import (
	"context"
	"fmt"
	"sort"
)

// Warning is a human-readable warning that indicates that there was a problem parsing a
// row in the spreadsheet
type Warning struct {
	// RowNumber is the user-facing row number (i.e. starting at 1 for the heading row,
	// 2 for the first tape) that indicates where the parsing error occurred
	RowNumber int
	// Message is the human-readable Message representing the error that occurred
	Message string
}

func ListTapes(ctx context.Context, c Client) ([]Tape, []Warning, error) {
	// Fetch the full contents of the Golden VCR Inventory spreadsheet's 'Tapes' sheet
	result, err := c.GetValues(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get values from inventory spreadsheet: %w", err)
	}

	// The first row contains column headings, with each row thereafter representing a
	// single tape: if the spreadsheet is entirely empty, consider it a fatal error
	if len(result.Values) == 0 {
		return nil, nil, fmt.Errorf("inventory spreadsheet has no values")
	}

	// Parse the headings in the first row to determine what column each value is
	// located in
	indexMap, err := newIndexMap(result.Values[0])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse headings from first row of inventory spreadsheet: %w", err)
	}

	// We want to be somewhat tolerant of malformed data - i.e. if we've entered some
	// placeholder data in the spreadsheet but we haven't finished populating the row,
	// we don't want that to cause the entire sync process to fail. Instead, we'll
	// collect non-fatal parsing errors into a list of warnings that can be surfaced to
	// the user.
	warnings := make([]Warning, 0)

	// Iterate through each row, attempting to parse the values in that row to a valid
	// Tape and logging a warning if unable: use a map so we can check for duplicate IDs
	tapesById := make(map[int]*Tape)
	idsWithMultipleTapes := make(map[int]struct{})
	for i := 1; i < len(result.Values); i++ {
		// If the row can't be parsed to a valid tape, log a warning and skip it
		values := rowValues(result.Values[i])
		tape, err := indexMap.parseRow(values)
		if err != nil {
			warnings = append(warnings, Warning{
				RowNumber: i + 1,
				Message:   err.Error(),
			})
			continue
		}

		// If we encounter a duplicate tape ID, skip this row and all other rows with
		// that same ID
		existing, found := tapesById[tape.Id]
		if found {
			warnings = append(warnings, Warning{
				RowNumber: i + 1,
				Message:   fmt.Sprintf("duplicate tape ID %d: used by both '%s' and '%s'; accepting neither", tape.Id, tape.Title, existing.Title),
			})
			// We can't remove the existing tape from the map while we're still
			// iterating, because that would allow a third tape with the same ID to be
			// accepted: instead, record the ID in a separate set so we can skip it when
			// building the final result array
			idsWithMultipleTapes[tape.Id] = struct{}{}
			continue
		}

		// Tape is good and ID is unique (so far)
		tapesById[tape.Id] = tape
	}

	// Build a final list of tapes to return to the caller, sorted by ID, skipping any
	// duplicates
	tapeIds := make([]int, 0, len(tapesById)-len(idsWithMultipleTapes))
	for tapeId := range tapesById {
		_, hasDuplicates := idsWithMultipleTapes[tapeId]
		if !hasDuplicates {
			tapeIds = append(tapeIds, tapeId)
		}
	}
	sort.Ints(tapeIds)
	tapes := make([]Tape, 0, len(tapeIds))
	for _, tapeId := range tapeIds {
		tapes = append(tapes, *tapesById[tapeId])
	}
	return tapes, warnings, nil
}
