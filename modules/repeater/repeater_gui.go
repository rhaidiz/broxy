package repeater

import (
	"strconv"
	"github.com/rhaidiz/broxy/core"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

var tabNum int

// Gui represents the Gui of the repeater module
type Gui struct {
	core.GuiModule
	Sess *core.Session

	repeaterTabs *widgets.QTabWidget
	tabs         []*TabGui
	tabNum       int
	tabRemoved   bool

	//GoClick func(*TabGui)
	GoClick func(int, string, string, chan string)
	NewTabEvent func(string, string)
	RemoveTabEvent func(*TabGui)
	Load func()
	_       func(i int) `signal:"changedTab"`
}

// Tab represents a tab in the repeater module
type TabGui struct {
	id				int
	goBtn          *widgets.QPushButton
	cancelBtn      *widgets.QPushButton
	HostLine       *widgets.QLineEdit
	RequestEditor  *widgets.QPlainTextEdit
	ResponseEditor *widgets.QPlainTextEdit
}

// NewGui creates a new Gui for the repeater module
func NewGui(s *core.Session) *Gui {
	tabNum = 1
	return &Gui{Sess: s, tabNum: 1, tabRemoved: false}
}

func (g *Gui) GetSettings() interface{} {
	return nil
}

// GetModuleGui returns the Gui for the current module
func (g *Gui) GetModuleGui() interface{}  {

	g.repeaterTabs = widgets.NewQTabWidget(nil)
	g.Load()
	g.repeaterTabs.SetDocumentMode(true)
	g.repeaterTabs.SetTabsClosable(true)
	g.repeaterTabs.ConnectTabCloseRequested(g.handleClose)
	g.repeaterTabs.ConnectCurrentChanged(g.changedTab)
	//g.repeaterTabs.AddTab(g.NewTab(), strconv.Itoa(g.tabNum))
	g.repeaterTabs.AddTab(widgets.NewQWidget(nil, 0), "+")
	// the following line is to remove the closable button from the last tab
	g.repeaterTabs.TabBar().SetTabButton(g.repeaterTabs.Count()-1, widgets.QTabBar__LeftSide, nil) //.Hide()

	return g.repeaterTabs

}

func (g *Gui) handleClose(index int) {
	g.tabRemoved = true
	g.RemoveTabEvent(g.tabs[index])
	g.repeaterTabs.RemoveTab(index)
}

func (g *Gui) changedTab(i int) {
	if i == g.repeaterTabs.Count()-1 && g.tabRemoved && g.repeaterTabs.Count() > 1 {
		g.repeaterTabs.SetCurrentIndex(i - 1)
	} else if i == g.repeaterTabs.Count()-1 {
		// This branch runs only when a new tab is added with the + button
		// or the first time I load the interface
		g.NewTabEvent("","")
	}
	g.tabRemoved = false
}

// AddNewTab adds a new repeater tab
func (g *Gui) AddNewTab(id int, host, request, response string) {
	g.repeaterTabs.InsertTab(g.repeaterTabs.Count()-1, g.NewTab(id, host, request, response), strconv.Itoa(id))
	g.repeaterTabs.SetCurrentIndex(g.repeaterTabs.Count() - 2)
}

// NewTab adds a new tab
func (g *Gui) NewTab(id int, host, request, response string) widgets.QWidget_ITF {
	t := &TabGui{id: id}
	g.tabs = append(g.tabs, t)
	mainWidget := widgets.NewQWidget(nil, 0)
	vlayout := widgets.NewQVBoxLayout()
	vlayout.SetContentsMargins(11, 11, 11, 11)
	mainWidget.SetLayout(vlayout)

	hlayout := widgets.NewQHBoxLayout()

	t.goBtn = widgets.NewQPushButton2("Go", nil)
	t.cancelBtn = widgets.NewQPushButton2("Cancel", nil)
	t.goBtn.ConnectClicked(func(b bool) {
		c := make(chan string)
		request := t.RequestEditor.ToPlainText()
		go g.GoClick(id, host, request, c)
		go func(){
			for resp := range c{
				t.ResponseEditor.SetPlainText(resp)
			}
		}()
	})
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
	t.ResponseEditor.SetPlainText(response)
	splitter.AddWidget(t.RequestEditor)
	splitter.AddWidget(t.ResponseEditor)

	vlayout.AddWidget(splitter, 0, 0)
	return mainWidget
}

// Title returns the time of this Gui
func (g *Gui) Title() string {
	return "Repeater"
}
