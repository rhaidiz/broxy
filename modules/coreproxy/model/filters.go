package model

// Filter represents the filter of the history table
type Filter struct {
	Search     string
	StatusCode []int
	Show			 bool
	Hide			 bool
	ShowExt    []string
	HideExt    []string
}


var DefaultFilter = & Filter{
	Search 			: "Test",
	StatusCode	: []int{100, 200, 300, 400, 500},
	Show				: false,
	Hide				: true,
	HideExt 		: []string{"jpg", "png"},
	ShowExt 		: []string{"asp","php"},

}
