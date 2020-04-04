package core

// Module interface
type Module interface {
	Name() string
	Description() string
	Status() bool
	Start() error
	Stop() error
}

// GuiModule interface
type GuiModule interface {
	GetModuleGui() interface{}
	GetSettings() interface{}
	Title() string
}

// ControllerModule interface
type ControllerModule interface {
	ExecCommand(string, ...interface{})
	GetModule() Module
	GetGui() GuiModule
}
