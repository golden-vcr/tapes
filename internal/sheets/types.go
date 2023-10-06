package sheets

// Tape represents a single tape from in the Golden VCR Inventory spreadsheet that has
// the minimal required information recorded to be included in the inventory
type Tape struct {
	// Unique ID of the tape, parsed from the ID column; must be set
	Id int
	// Title of the tape; must be set
	Title string
	// Publication year of the tape as an integer, or 0 if unknown
	Year int
	// Approximate runtime of the tape in minutes, or 0 if unknown
	Runtime int
}
