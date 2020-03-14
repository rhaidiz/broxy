package main

import (
	"os"

	"github.com/rhaidiz/broxy/gui"
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/util"
	"github.com/therecipe/qt/widgets"
)

func main() {

	qa := widgets.NewQApplication(len(os.Args), os.Args)
	cfg := core.LoadGlobalSettings(util.GetSettingsDir())
	history := core.LoadHistory(util.GetSettingsDir())
	prj := gui.NewProjectgui(nil, 0)
	prj.InitWith(history, cfg, qa)

	prj.Show()

	widgets.QApplication_Exec()
}
