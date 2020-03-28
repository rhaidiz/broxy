package core

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// Broxygui is the main GUI made of tabs
type Broxygui struct {
	widgets.QMainWindow
	_ func() `constructor:"setup"`

	tabWidget *widgets.QTabWidget
	treeWidget *widgets.QTreeWidget

	settingsMapping 			map[string]widgets.QWidget_ITF
	modulesTreeItem 			*widgets.QTreeWidgetItem
	current 					string
	hLayout 					*widgets.QHBoxLayout
	gzipDecodeCheckBox          *widgets.QCheckBox

	s *Session
}

func (g *Broxygui) setup() {
	// loading global config

	g.settingsMapping = make(map[string]widgets.QWidget_ITF)
	g.SetWindowTitle("Broxy (1.0.0-alpha.2)")
	//g.SetMinimumSize(core.NewQSize2(523, 317))

	g.tabWidget = widgets.NewQTabWidget(nil)
	g.tabWidget.SetDocumentMode(true)

	g.SetCentralWidget(g.tabWidget)
	g.tabWidget.AddTab(g.settingsTab(), "Settings")
}

func (g *Broxygui) InitWith(s *Session) {
	g.s = s
	if s.GlobalSettings.GZipDecode {
		g.gzipDecodeCheckBox.SetChecked(true)
	}else{
		g.gzipDecodeCheckBox.SetChecked(false)
	}
	
}

//AddGuiModule adds a new module to the main GUI
func (g *Broxygui) AddGuiModule(m GuiModule) {
	g.tabWidget.SetCurrentIndex(0)
	g.tabWidget.InsertTab(0,m.GetModuleGui(), m.Title())
	if m.GetSettings() != nil {
		g.settingsMapping[m.Title()] = m.GetSettings()
		item := widgets.NewQTreeWidgetItem(0)
		item.SetText(0,m.Title())
		g.modulesTreeItem.AddChild(item)
		g.modulesTreeItem.SetExpanded(true)
	}
}

//ShowErrorMessage shows a critical message box
func (g *Broxygui) ShowErrorMessage(message string) {
	widgets.QMessageBox_Critical(nil, "OK", message, widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
}

func (g *Broxygui) settingsTab() widgets.QWidget_ITF{
	widget := widgets.NewQWidget(nil, 0)
	g.hLayout = widgets.NewQHBoxLayout()
	widget.SetLayout(g.hLayout)

	g.treeWidget = widgets.NewQTreeWidget(nil)
	g.treeWidget.ConnectItemClicked(g.itemClicked)
	g.treeWidget.SetHeaderHidden(true)
	g.hLayout.AddWidget(g.treeWidget,0 ,0)

	item := widgets.NewQTreeWidgetItem(0)
	item.SetText(0,"Global Settings")

	g.modulesTreeItem = widgets.NewQTreeWidgetItem(0)
	g.modulesTreeItem.SetText(0, "Modules")

	g.treeWidget.AddTopLevelItem(item)
	g.treeWidget.AddTopLevelItem(g.modulesTreeItem)
	//g.treeWidget.SetSizePolicy(widgets.QSizePolicy__Fixed)
	g.treeWidget.SetFixedWidth(200)

	g.treeWidget.SetCurrentItem(item)
	global := g.globalSettings()
	g.hLayout.AddWidget(global,0 ,0)

	g.current = "Global Settings"
	g.settingsMapping["Global Settings"] = global
	//g.settingsMapping["Modules"] = g.emptySettings()

	return widget
}



func (g *Broxygui) globalSettings() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	hLayout := widgets.NewQVBoxLayout()
	widget.SetLayout(hLayout)

	label := widgets.NewQLabel(nil, 0)
	font := gui.NewQFont()
	font.SetPointSize(20)
	font.SetBold(true)
	font.SetWeight(75)
	label.SetFont(font)
	label.SetObjectName("label")
	label.SetText("Global Settings")

	g.gzipDecodeCheckBox = widgets.NewQCheckBox(nil)
	g.gzipDecodeCheckBox.SetText("Decode GZIP Responses")
	g.gzipDecodeCheckBox.ConnectClicked(g.gzipDecodeCheckBoxClicked)

	spacerItem := widgets.NewQSpacerItem(20, 40, widgets.QSizePolicy__Minimum, widgets.QSizePolicy__Expanding)

	hLayout.AddWidget(label, 0, core.Qt__AlignLeft)
	hLayout.AddWidget(g.gzipDecodeCheckBox, 0, core.Qt__AlignLeft)
	hLayout.AddItem(spacerItem)
	
	return widget
}

func ( g *Broxygui) gzipDecodeCheckBoxClicked(b bool){
	g.s.GlobalSettings.GZipDecode = g.gzipDecodeCheckBox.IsChecked()
	g.s.PersistentProject.SaveSettings("project",g.s.GlobalSettings)
}

func (g *Broxygui) emptySettings() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	hLayout := widgets.NewQHBoxLayout()
	widget.SetLayout(hLayout)
	hLayout.AddWidget(widgets.NewQPushButton2("AAAAAA", nil),0,0)
	return widget
}

func (g *Broxygui) itemClicked(item *widgets.QTreeWidgetItem, column int){
	if _, ok := g.settingsMapping[item.Text(0)]; ok {
		g.hLayout.ReplaceWidget(g.settingsMapping[g.current], g.settingsMapping[item.Text(0)], core.Qt__FindChildrenRecursively)
		g.settingsMapping[g.current].QWidget_PTR().SetVisible(false)
		g.settingsMapping[item.Text(0)].QWidget_PTR().SetVisible(true)
		g.current = item.Text(0)
	}
}
