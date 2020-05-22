package core

import (
	"time"
	"github.com/rhaidiz/broxy/core/project"
	"github.com/rhaidiz/broxy/util"
)

type MainGui interface {
	AddGuiModule(GuiModule)
	InitWith(*Session)
	ShowErrorMessage(string)
}

// Session represents a running session in Broxy with a GUI and loaded modules
type Session struct {
	Controllers 		[]ControllerModule
	
	Logs 				[]Log
	LogEvent			chan Log

	MainGui 			MainGui
	PersistentProject	*project.PersistentProject

	Settings  			*BroxySettings
	GlobalSettings 		*GlobalSettings
}

// NewSession creates a new session
func NewSession(cfg *BroxySettings, p *project.PersistentProject, gui MainGui) *Session {
	gc := &GlobalSettings{}
	p.LoadSettings("project",gc)
	s := &Session{
		MainGui: 			gui,
		Settings:  			cfg,
		GlobalSettings: 	gc,
		LogEvent:    		make(chan Log),
		PersistentProject:	p,
	}
	gui.InitWith(s)
	return s
}

// LoadModule loads a module in the current session
func (s *Session) LoadModule(c ControllerModule) {
	if !util.IsNil(c) {
		s.Controllers = append(s.Controllers, c)
		s.MainGui.AddGuiModule(c.GetGui())
	}
}

// Exec executes, for a given module m, a function f with parameters a
func (s *Session) Exec(c string, f string, a ...interface{}) {
	for _, ctrl := range s.Controllers {
		println(ctrl.GetModule().Name())
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
	go func() { s.LogEvent <- l }()
}

// Debug logs a debug information messasge in the current session
func (s *Session) Debug(mod string, message string) {
	t := time.Now()
	l := Log{Type: "D", ModuleName: mod, Time: t.Format("2006-01-02 15:04:05"), Message: message}
	s.Logs = append(s.Logs, l)
	go func() { s.LogEvent <- l }()
}

// Err logs an error information message in the current session
func (s *Session) Err(mod string, message string) {
	t := time.Now()
	l := Log{Type: "E", ModuleName: mod, Time: t.Format("2006-01-02 15:04:05"), Message: message}
	s.Logs = append(s.Logs, l)
	go func() { s.LogEvent <- l }()
}

func (s *Session) ShowErrorMessage(message string){
	s.MainGui.ShowErrorMessage(message)
}
