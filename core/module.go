package core

import (
	"github.com/therecipe/qt/widgets"
)

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
	GetModuleGui() widgets.QWidget_ITF
	Title() string
}

// ControllerModule interface
type ControllerModule interface {
	ExecCommand(string, ...interface{})
	GetModule() Module
	GetGui() GuiModule
}
