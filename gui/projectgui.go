package gui

import (
	"fmt"
	_ "os"
	bcore "github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules"
	"github.com/rhaidiz/broxy/util"
	"github.com/rhaidiz/broxy/core/project"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"time"
	"path/filepath"
)

// Projectgui shows a project history and allows to create a new project or load an existing one
type Projectgui struct {
	widgets.QMainWindow
	_ func() `constructor:"setup"`

	projectsListWidget *widgets.QListWidget

	newProjectButton  *widgets.QPushButton
	loadProjectButton *widgets.QPushButton
	openProjectButton *widgets.QPushButton
	qApp              *widgets.QApplication
	config            *bcore.BroxySettings
	history 		  *project.History
	contextMenu       *widgets.QMenu


}

func (g *Projectgui) setup() {

}

// InitWith initializes Projectgui with a given history, configuration and QApplication
func (g *Projectgui) InitWith(qApp *widgets.QApplication) {
	g.qApp = qApp
	g.init()
}

func (g *Projectgui) init(){

	g.config = bcore.LoadGlobalSettings(util.GetSettingsDir())
	g.history = project.LoadHistory(util.GetSettingsDir())

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

	delegate := InitDelegate(g.qApp)
	g.projectsListWidget.SetItemDelegate(delegate)
	g.projectsListWidget.ConnectItemDoubleClicked(g.itemDoubleClicked)
	g.projectsListWidget.SetContextMenuPolicy(core.Qt__CustomContextMenu)
	g.projectsListWidget.ConnectCustomContextMenuRequested(g.customContextMenuRequested)

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

func (g *Projectgui) customContextMenuRequested(p *core.QPoint) {
	if g.contextMenu == nil {
		g.contextMenu = widgets.NewQMenu(nil)
		remove := g.contextMenu.AddAction("Remove")
		remove.ConnectTriggered(func(b bool) {
			fmt.Println("right click")
			r := g.projectsListWidget.CurrentRow()
			g.history.Remove(g.history.H[r])
			g.projectsListWidget.TakeItem(r)
		})
	}
	g.contextMenu.Exec2(g.projectsListWidget.MapToGlobal(p), nil)
}

func (g *Projectgui) itemDoubleClicked(item *widgets.QListWidgetItem){
	r := g.projectsListWidget.CurrentRow()
	path := g.history.H[r].Path
	title := g.history.H[r].Title
	c, err := project.OpenPersistentProject(title,path)
	if err != nil {
		g.showErrorMessage("Project does not exist")
		return
	}
	gui := NewBroxygui(nil,0)
	s := bcore.NewSession(g.config, c, gui)
	//Load All modules
	modules.LoadModules(s)

	gui.Show()
	g.Close()
}

func (g *Projectgui) newProject(b bool) {

	p := filepath.Join(util.GetTmpDir(), fmt.Sprintf("%d",time.Now().UnixNano()))
	fmt.Println(p)
	c, _ := project.NewPersistentProject("NewProject",p)

	// temporary, for now, everytime I create a new project I save it in the history
	gui := NewBroxygui(nil,0)
	s := bcore.NewSession(g.config, c, gui)
	//Load All modules
	modules.LoadModules(s)

	g.history.Add(&project.Project{"NewProject",p})
	gui.Show()
	g.Close()
}

//ShowErrorMessage shows a critical message box
func (g *Projectgui) showErrorMessage(message string) {
	widgets.QMessageBox_Critical(nil, "OK", message, widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
}