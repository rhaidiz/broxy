package core

import "fmt"

type Log struct {
	Type       string
	ModuleName string
	Time       string
	Message    string
}

func (l *Log) ToString() string {
	return fmt.Sprintf("[%s][%s][%s] %s", l.Type, l.ModuleName, l.Time, l.Message)
}
