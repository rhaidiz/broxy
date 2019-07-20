package core

type Session struct {

	// represent the session on FS
	Path string

	// List of modules
	Modules []Module

	MainGui *Broxygui
	Config  *Config
}

func NewSession(path string) *Session {
	return &Session{
		Path:    path,
		MainGui: NewBroxygui(nil, 0),
		Config: &Config{
			Address: "127.0.0.1",
			Port:    8080,
		},
	}
}

func LoadSession(path string) *Session {
	// Load session from file
	return nil
}

func (s *Session) LoadModule(m Module, g GuiModule) {
	s.Modules = append(s.Modules, m)
	s.MainGui.AddGuiModule(g)
}
