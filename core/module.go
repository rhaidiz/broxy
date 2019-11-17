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

type GuiModule interface {
	GetModuleGui() widgets.QWidget_ITF
	Name() string
}

type ControllerModule interface {
	ExecCommand(string, ...interface{})
	GetModule() Module
	GetGui() GuiModule
	Name() string
}
