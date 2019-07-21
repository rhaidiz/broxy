package log

import (
	"github.com/rhaidiz/broxy/core"
)

type Log struct {
	core.Module
}

//var mutex = &sync.Mutex{}

// Create a new proxy
func NewLog(s *core.Session) *Log {
	// this is my struct that I use to represent the proxy
	return &Log{}
}

func (m *Log) Name() string {
	return "Log"
}
