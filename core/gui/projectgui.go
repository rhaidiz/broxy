package coregui

import (
	"fmt"
	bcore "github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/util"
	"github.com/rhaidiz/broxy/modules"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"time"
	"path/filepath"
)

type Projectgui struct {
	widgets.QMainWindow
	_ func() `constructor:"setup"`

	projectsListWidget *widgets.QListWidget

	newProjectButton  *widgets.QPushButton
	loadProjectButton *widgets.QPushButton
	openProjectButton *widgets.QPushButton
	QApp              *widgets.QApplication
	Config            *bcore.Config
	history 					*bcore.History
}

func (g *Projectgui) setup() {


}


func (g *Projectgui) InitWith(history *bcore.History, cfg *bcore.Config, qApp *widgets.QApplication) {
	g.Config = cfg
	g.QApp = qApp
	g.history = history
	g.init()
}

func (g *Projectgui) init(){
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
	g.projectsListWidget.ConnectItemDoubleClicked(g.itemDoubleClicked)
	font := gui.NewQFont2("Monospace", 11, int(gui.QFont__Normal), false)
	fontMetrics := gui.NewQFontMetricsF(font)

	p := ""
	t := ""
	for _,h := range g.history.H {
	//println("first instruction in loop")
		p = h.Path
		// this ElidedText is performing very poorly, I might implement something myself
		localElidedText := fontMetrics.ElidedText(p, core.Qt__ElideMiddle, 230, 0)
		t = fmt.Sprintf("<big>%s</big><br>", h.Title)
		g.projectsListWidget.AddItem(t + localElidedText)
	}


	hLayout.AddWidget(g.projectsListWidget, 0, 0)

	rightWidget := widgets.NewQWidget(nil, 0)
	vLayout := widgets.NewQVBoxLayout()
	rightWidget.SetLayout(vLayout)

	hLayout.AddWidget(rightWidget, 0, 0)

	g.newProjectButton = widgets.NewQPushButton2("New Project", nil)
	g.newProjectButton.ConnectClicked(g.newProject)

	//g.loadProjectButton = widgets.NewQPushButton2("Load Project", nil)
	g.openProjectButton = widgets.NewQPushButton2("Open Existing Project", nil)

	spacerItem := widgets.NewQSpacerItem(40, 20, widgets.QSizePolicy__Minimum, widgets.QSizePolicy__Expanding)

	vLayout.AddItem(spacerItem)
	vLayout.AddWidget(g.newProjectButton, 0, 0)
	//vLayout.AddWidget(g.loadProjectButton, 0, 0)
	vLayout.AddWidget(g.openProjectButton, 0, 0)
	vLayout.AddItem(spacerItem)
}

func (g *Projectgui) itemDoubleClicked(item *widgets.QListWidgetItem){
	r := g.projectsListWidget.CurrentRow()
	
	g.Config.Project = g.history.H[r]
	s := bcore.NewSession(g.QApp, g.Config)
	//Load All modules
	modules.LoadModules(s)

	s.MainGui.Show()
	g.Close()
}

func (g *Projectgui) newProject(b bool) {

	p := filepath.Join(util.GetTmpDir(), fmt.Sprintf("%d",time.Now().UnixNano()))
	prj := &bcore.Project{"New empty project",p}
	g.Config.Project = prj
	s := bcore.NewSession(g.QApp, g.Config)
	//Load All modules
	modules.LoadModules(s)

	s.MainGui.Show()
	g.Close()
}
