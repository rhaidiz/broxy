package core

import (
	_ "encoding/pem"
	_ "encoding/xml"
	_ "fmt"
	_ "io/ioutil"
	_ "os"
	_ "path/filepath"
	"time"

	"github.com/therecipe/qt/widgets"
)


// Session represents a running session in Broxy with a GUI and loaded modules
type Session struct {


	// List of modules
	Controllers []ControllerModule

	// Logs
	Logs []Log

	MainGui *Broxygui
	Config  *Config

	LogC chan Log

	QApp *widgets.QApplication
}

// NewSession creates a new session
func NewSession(qa *widgets.QApplication, cfg *Config) *Session {

	return &Session{
		MainGui: NewBroxygui(nil, 0),
		Config:  cfg,
		LogC:    make(chan Log),
		QApp:    qa,
	}
}

// LoadModule loads a module in the current session
func (s *Session) LoadModule(c ControllerModule) {
	s.Controllers = append(s.Controllers, c)
	s.MainGui.AddGuiModule(c.GetGui())
}

// Exec executes, for a given module m, a function f with parameters a
func (s *Session) Exec(c string, f string, a ...interface{}) {
	for _, ctrl := range s.Controllers {
		if c == ctrl.GetModule().Name() {
			ctrl.ExecCommand(f, a...)
		}
	}
}

// Info logs an information message in the current session
func (s *Session) Info(mod string, message string) {
	t := time.Now()
	l := Log{Type: "I", ModuleName: mod, Time: t.Format("2006-01-02 15:04:05"), Message: message}
	s.Logs = append(s.Logs, l)
	go func() { s.LogC <- l }()
}

// Debug logs a debug information messasge in the current session
func (s *Session) Debug(mod string, message string) {
	t := time.Now()
	l := Log{Type: "D", ModuleName: mod, Time: t.Format("2006-01-02 15:04:05"), Message: message}
	s.Logs = append(s.Logs, l)
	go func() { s.LogC <- l }()
}

// Err logs an error information message in the current session
func (s *Session) Err(mod string, message string) {
	t := time.Now()
	l := Log{Type: "E", ModuleName: mod, Time: t.Format("2006-01-02 15:04:05"), Message: message}
	s.Logs = append(s.Logs, l)
	go func() { s.LogC <- l }()
}
