package gui

import (
	"os"
	"path/filepath"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"github.com/rhaidiz/broxy/core/project"
)

// History represents a history of projects
type History struct {
	H []*project.Project `json:"ProjectsHistory"`
	path string
}

var historyFileName = "history.json"

// LoadHistory loads the projects history stored in path
func LoadHistory(path string) *History {

	historyPath := filepath.Join(path, historyFileName)
	historyFile, err := os.Open(historyPath)
	defer historyFile.Close()
	if err != nil {
		// no history
		return &History{path:path}
	}
	byteValue, err := ioutil.ReadAll(historyFile)
	// TODO: handle error
	if err != nil {
		fmt.Println(err)
	}
	var history *History
	err = json.Unmarshal(byteValue, &history)
	if err != nil {
		return &History{path:path}
	}
	history.path = path
	return history
}

// SaveToHistory saves a new project to the history
func (h *History) Add(p *project.Project) error {
	h.H = append(h.H, p)

	historyJson, _ := json.MarshalIndent(h, "", " ")
	historyFile := filepath.Join(h.path, historyFileName)

	return ioutil.WriteFile(historyFile, historyJson, 0700)
}

// RemoveFromHistory removes an entry from the history
func (h *History) Remove(p *project.Project) error{
	r := -1
	for i,e := range h.H {
		if e.Title == p.Title && e.Path == p.Path{
			r = i
		}
	}
	copy(h.H[r:], h.H[r+1:]) // Shift a[i+1:] left one index.
	h.H = h.H[:len(h.H)-1]     // Truncate slice.

	historyJson, _ := json.MarshalIndent(h, "", " ")
	historyFile := filepath.Join(h.path, historyFileName)

	return ioutil.WriteFile(historyFile, historyJson, 0700)
}