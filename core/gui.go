package core

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

// this will load the main GUI which is made of tabs

type Broxygui struct {
	widgets.QMainWindow
	_ func() `constructor:"setup"`

	tabWidget *widgets.QTabWidget
}

func (g *Broxygui) setup() {

	g.SetWindowTitle("Broxy (Beta)")
	g.SetMinimumSize(core.NewQSize2(523, 317))

	g.tabWidget = widgets.NewQTabWidget(nil)
	g.tabWidget.SetDocumentMode(true)

	g.SetCentralWidget(g.tabWidget)
}

func (g *Broxygui) AddGuiModule(m GuiModule) {
	g.tabWidget.AddTab(m.GetModuleGui(), m.Name())
}

func (g *Broxygui) ShowErrorMessage(message string) {
	widgets.QMessageBox_Critical(nil, "OK", message, widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
}
