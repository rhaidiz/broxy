package main

import (
	"os"

	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules"
	"github.com/rhaidiz/broxy/util"
	"github.com/therecipe/qt/widgets"
)

func main() {

	qa := widgets.NewQApplication(len(os.Args), os.Args)

	s := core.NewSession(util.GetSettingsDir(), qa)
	//Load All modules
	modules.LoadModules(s)

	s.MainGui.Show()

	widgets.QApplication_Exec()

}
