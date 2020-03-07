package main

import (
	"os"

	"github.com/rhaidiz/broxy/core"
	_ "github.com/rhaidiz/broxy/modules"
	_ "github.com/rhaidiz/broxy/util"
	"github.com/therecipe/qt/widgets"
)

func main() {

	qa := widgets.NewQApplication(len(os.Args), os.Args)

	prj:= core.NewProjectgui(nil, 0)
	prj.QApp = qa

	prj.Show()

	// config := core.LoadGlobalSettings(util.GetSettingsDir())
	// s := core.NewSession("", qa, config)
	// //Load All modules
	// modules.LoadModules(s)

	// s.MainGui.Show()

	widgets.QApplication_Exec()

}
