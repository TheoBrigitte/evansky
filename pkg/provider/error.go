package provider

import "fmt"

var (
	// ErrNoResults is returned when no results are found.
	ErrNoResult = fmt.Errorf("no result found")
)
