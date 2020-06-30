package main

import (
	"os"

	"github.com/rhaidiz/broxy/gui"
	"github.com/therecipe/qt/widgets"
)

func main() {

	qa := widgets.NewQApplication(len(os.Args), os.Args)
	prj := gui.NewProjectgui(nil, 0)
	prj.InitWith(qa)

	prj.Show()

	widgets.QApplication_Exec()
}
