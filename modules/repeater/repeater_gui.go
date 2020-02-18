package repeater

import (
	"strconv"

	"github.com/rhaidiz/broxy/core"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

// RepeaterGui represents the Gui of the repeater module
type RepeaterGui struct {
	core.GuiModule
	Sess *core.Session

	repeaterTabs *widgets.QTabWidget
	tabs         []*RepeaterTab
	tabNum       int
	tabRemoved   bool

	GoClick func(*RepeaterTab)
	_       func(i int) `signal:"changedTab"`
}

// RepeaterTab represents a tab in the repeater module
type RepeaterTab struct {
	goBtn          *widgets.QPushButton
	cancelBtn      *widgets.QPushButton
	HostLine       *widgets.QLineEdit
	RequestEditor  *widgets.QPlainTextEdit
	ResponseEditor *widgets.QPlainTextEdit
}

// NewRepeaterGui creates a new Gui for the repeater module
func NewRepeaterGui(s *core.Session) *RepeaterGui {
	return &RepeaterGui{Sess: s, tabNum: 1, tabRemoved: false}
}

// GetModuleGui returns the Gui for the current module
func (g *RepeaterGui) GetModuleGui() widgets.QWidget_ITF {

	g.repeaterTabs = widgets.NewQTabWidget(nil)
	g.repeaterTabs.SetDocumentMode(true)
	g.repeaterTabs.ConnectCurrentChanged(g.changedTab)
	g.repeaterTabs.SetTabsClosable(true)
	g.repeaterTabs.ConnectTabCloseRequested(g.handleClose)

	//g.repeaterTabs.AddTab(g.NewTab(), strconv.Itoa(g.tabNum))
	g.repeaterTabs.AddTab(g.NewEmptyTab(), "+")

	return g.repeaterTabs

}

func (g *RepeaterGui) handleClose(index int) {
	g.tabRemoved = true
	g.repeaterTabs.RemoveTab(index)
}

func (g *RepeaterGui) changedTab(i int) {
	if i == g.repeaterTabs.Count()-1 && g.tabRemoved && g.repeaterTabs.Count() > 1 {
		g.repeaterTabs.SetCurrentIndex(i - 1)
	} else if i == g.repeaterTabs.Count()-1 {
		// add a new tab before me
		g.repeaterTabs.InsertTab(g.repeaterTabs.Count()-1, g.NewEmptyTab(), strconv.Itoa(g.tabNum))
		g.tabNum = g.tabNum + 1
		g.repeaterTabs.SetCurrentIndex(g.repeaterTabs.Count() - 2)
	}
	g.tabRemoved = false
}

// AddNewTab adds a new repeater tab
func (g *RepeaterGui) AddNewTab(host string, request string) {
	g.tabNum = g.tabNum + 1
	g.repeaterTabs.InsertTab(g.repeaterTabs.Count()-1, g.NewTab(host, request), host)
	g.repeaterTabs.SetCurrentIndex(g.repeaterTabs.Count() - 2)
}

// NewTab adds a new tab
func (g *RepeaterGui) NewTab(host string, request string) widgets.QWidget_ITF {
	t := &RepeaterTab{}
	g.tabs = append(g.tabs, t)
	mainWidget := widgets.NewQWidget(nil, 0)
	vlayout := widgets.NewQVBoxLayout()
	vlayout.SetContentsMargins(11, 11, 11, 11)
	mainWidget.SetLayout(vlayout)

	hlayout := widgets.NewQHBoxLayout()

	t.goBtn = widgets.NewQPushButton2("Go", nil)
	t.cancelBtn = widgets.NewQPushButton2("Cancel", nil)
	t.goBtn.ConnectClicked(func(b bool) { g.GoClick(t) })
	hlayout.AddWidget(t.goBtn, 0, 0)
	hlayout.AddWidget(t.cancelBtn, 0, 0)

	t.HostLine = widgets.NewQLineEdit(nil)
	t.HostLine.SetText(host)
	hlayout.AddWidget(t.HostLine, 0, 0)

	vlayout.AddLayout(hlayout, 0)

	splitter := widgets.NewQSplitter(nil)
	splitter.SetOrientation(qtcore.Qt__Horizontal)

	t.RequestEditor = widgets.NewQPlainTextEdit(nil)
	t.RequestEditor.SetPlainText(request)
	t.ResponseEditor = widgets.NewQPlainTextEdit(nil)
	t.ResponseEditor.SetReadOnly(true)
	splitter.AddWidget(t.RequestEditor)
	splitter.AddWidget(t.ResponseEditor)

	vlayout.AddWidget(splitter, 0, 0)

	return mainWidget
}

// NewEmptyTab adds a new empty tab
func (g *RepeaterGui) NewEmptyTab() widgets.QWidget_ITF {
	return g.NewTab("", "")

}

// Title returns the time of this Gui
func (g *RepeaterGui) Title() string {
	return "Repeater"
}
