package main

import (
	"os"

	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/core/gui"
	_ "github.com/rhaidiz/broxy/modules"
	"github.com/rhaidiz/broxy/util"
	"github.com/therecipe/qt/widgets"
)

func main() {

	qa := widgets.NewQApplication(len(os.Args), os.Args)
	cfg := core.LoadGlobalSettings(util.GetSettingsDir())
	prj := coregui.NewProjectgui(nil, 0)
	prj.InitWith(cfg, qa)

	prj.Show()

	widgets.QApplication_Exec()

}
