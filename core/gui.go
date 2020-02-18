package core

import (
	// "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

// Broxygui is the main GUI made of tabs
type Broxygui struct {
	widgets.QMainWindow
	_ func() `constructor:"setup"`

	tabWidget *widgets.QTabWidget
}

func (g *Broxygui) setup() {

	g.SetWindowTitle("Broxy (1.0.0-alpha.2)")
	//g.SetMinimumSize(core.NewQSize2(523, 317))

	g.tabWidget = widgets.NewQTabWidget(nil)
	g.tabWidget.SetDocumentMode(true)

	g.SetCentralWidget(g.tabWidget)
}

//AddGuiModule adds a new module to the main GUI
func (g *Broxygui) AddGuiModule(m GuiModule) {
	g.tabWidget.AddTab(m.GetModuleGui(), m.Title())
}

//ShowErrorMessage shows a critical message box
func (g *Broxygui) ShowErrorMessage(message string) {
	widgets.QMessageBox_Critical(nil, "OK", message, widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
}
