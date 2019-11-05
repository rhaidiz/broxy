package model

type Filter struct {
	Search     string
	StatusCode []int
	Show_ext   map[string]bool
	Hide_ext   map[string]bool
}
