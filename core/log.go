package core

import "fmt"

// Log represents a log message
type Log struct {
	Type       string
	ModuleName string
	Time       string
	Message    string
}

// ToString prints a string representation of a log message
func (l *Log) ToString() string {
	return fmt.Sprintf("[%s][%s][%s] %s", l.Type, l.ModuleName, l.Time, l.Message)
}
