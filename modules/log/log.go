package log

import (
	"github.com/rhaidiz/broxy/core"
)

// Log represents the log module
type Log struct {
	core.Module
}

//var mutex = &sync.Mutex{}

// NewLog creates a new log module
func NewLog(s *core.Session) *Log {
	return &Log{}
}

// Name returns the name of the current module
func (m *Log) Name() string {
	return "Log"
}
