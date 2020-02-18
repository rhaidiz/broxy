package core

import (
	"time"

	"github.com/therecipe/qt/widgets"
)

// Session represents a running session in Broxy with a GUI and loaded modules
type Session struct {

	// represent the session on FS
	Path string

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
func NewSession(path string, qa *widgets.QApplication) *Session {
	return &Session{
		Path:    path,
		MainGui: NewBroxygui(nil, 0),
		Config: &Config{
			Address:       "127.0.0.1",
			Port:          8080,
			ReqIntercept:  true,
			RespIntercept: false,
			Interceptor:   false,
		},
		LogC: make(chan Log),
		QApp: qa,
	}
}

// LoadSession loads a session from a path
// TODO: implement me
func LoadSession(path string) *Session {
	// Load session from file
	return nil
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
