package core

import (
	"github.com/therecipe/qt/widgets"
	"time"
)

type Session struct {

	// represent the session on FS
	Path string

	// List of modules
	Modules []ControllerModule

	// Logs
	Logs []Log

	MainGui *Broxygui
	Config  *Config

	LogC chan Log

	QApp *widgets.QApplication
}

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

func LoadSession(path string) *Session {
	// Load session from file
	return nil
}

func (s *Session) LoadModule(c ControllerModule) {
	s.Modules = append(s.Modules, c)
	s.MainGui.AddGuiModule(c.GetGui())
}

func (s *Session) Exec(m string, f string, a ...interface{}) {
	print(m)
	for _, mod := range s.Modules {
		if m == mod.Name() {
			mod.ExecCommand(f, a...)
		}
	}
}

func (s *Session) Info(mod string, message string) {
	t := time.Now()
	l := Log{Type: "I", ModuleName: mod, Time: t.Format("2006-01-02 15:04:05"), Message: message}
	s.Logs = append(s.Logs, l)
	go func() { s.LogC <- l }()
}

func (s *Session) Debug(mod string, message string) {
	t := time.Now()
	l := Log{Type: "D", ModuleName: mod, Time: t.Format("2006-01-02 15:04:05"), Message: message}
	s.Logs = append(s.Logs, l)
	go func() { s.LogC <- l }()
}

func (s *Session) Err(mod string, message string) {
	t := time.Now()
	l := Log{Type: "E", ModuleName: mod, Time: t.Format("2006-01-02 15:04:05"), Message: message}
	s.Logs = append(s.Logs, l)
	go func() { s.LogC <- l }()
}
