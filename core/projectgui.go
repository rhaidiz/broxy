package core

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

type Projectgui struct {
	widgets.QMainWindow
	_ func() `constructor:"setup"`

	projectsListWidget *widgets.QListWidget

	newProjectButton  *widgets.QPushButton
	loadProjectButton *widgets.QPushButton
	openProjectButton *widgets.QPushButton
	QApp *widgets.QApplication
}

func (g *Projectgui) setup() {
	g.SetWindowTitle("Welcome to Broxy")
	g.Resize(core.NewQSize2(488, 372))
	g.SetMinimumSize(core.NewQSize2(488, 372))
	g.SetMaximumSize(core.NewQSize2(488, 372))

	mainWidget := widgets.NewQWidget(nil, 0)
	hLayout := widgets.NewQHBoxLayout()
	hLayout.SetContentsMargins(0, 0, 12, 0)
	mainWidget.SetLayout(hLayout)
	g.SetCentralWidget(mainWidget)
	g.projectsListWidget = widgets.NewQListWidget(nil)

	delegate := InitDelegate(g.QApp)
	g.projectsListWidget.SetItemDelegate(delegate)
	//g.projectsListWidget.SetTextElideMode(core.Qt__ElideRight)
	font := gui.NewQFont2("Monospace", 11, int(gui.QFont__Normal), false)
	fontMetrics := gui.NewQFontMetricsF(font)
	t := "/path/to/project/very/very/very/very/very/very/very/long"
	localElidedText := fontMetrics.ElidedText(t, core.Qt__ElideMiddle, 230, 0)

	for i := 0; i < 30; i++ {
		g.projectsListWidget.AddItem("<big>Title</big><br>"+localElidedText)
		//g.projectsListWidget.AddItem("<big>Title</big><br>/path/to/project")
		//g.projectsListWidget.AddItem("<big>Title</big><br>/path/to/project/very/very/ver7very/very/very/very/long")
	}

	hLayout.AddWidget(g.projectsListWidget, 0, 0)

	rightWidget := widgets.NewQWidget(nil, 0)
	vLayout := widgets.NewQVBoxLayout()
	rightWidget.SetLayout(vLayout)

	hLayout.AddWidget(rightWidget, 0, 0)

	g.newProjectButton = widgets.NewQPushButton2("New Project", nil)
	g.loadProjectButton = widgets.NewQPushButton2("Load Project", nil)
	g.openProjectButton = widgets.NewQPushButton2("Open Existing Project", nil)

	spacerItem := widgets.NewQSpacerItem(40, 20, widgets.QSizePolicy__Minimum, widgets.QSizePolicy__Expanding)

	vLayout.AddItem(spacerItem)
	vLayout.AddWidget(g.newProjectButton, 0, 0)
	vLayout.AddWidget(g.loadProjectButton, 0, 0)
	vLayout.AddWidget(g.openProjectButton, 0, 0)
	vLayout.AddItem(spacerItem)

}

func (g *Projectgui) gui() {

}
