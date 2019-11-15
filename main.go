package main

import (
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules"
	"github.com/therecipe/qt/widgets"
	"os"
)

func main() {

	qa := widgets.NewQApplication(len(os.Args), os.Args)

	s := core.NewSession("~/Desktop", qa)
	//Load All modules
	modules.LoadModules(s)

	s.MainGui.Show()

	widgets.QApplication_Exec()

}
