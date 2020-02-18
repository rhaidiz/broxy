package model

// Filter represents the filter of the history table
type Filter struct {
	Search     string
	StatusCode []int
	ShowExt    map[string]bool
	HideExt    map[string]bool
}
