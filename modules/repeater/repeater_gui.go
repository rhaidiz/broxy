package repeater

import (
	"github.com/rhaidiz/broxy/core"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
	"strconv"
)

type RepeaterGui struct {
	core.GuiModule
	Sess *core.Session

	repeaterTabs *widgets.QTabWidget
	tabs         []*RepeaterTab
	tabNum       int

	GoClick func(*RepeaterTab)
	_       func(i int) `signal:"changedTab"`
}

type RepeaterTab struct {
	goBtn          *widgets.QPushButton
	cancelBtn      *widgets.QPushButton
	HostLine       *widgets.QLineEdit
	RequestEditor  *widgets.QPlainTextEdit
	ResponseEditor *widgets.QPlainTextEdit
}

func NewRepeaterGui(s *core.Session) *RepeaterGui {
	return &RepeaterGui{Sess: s, tabNum: 1}
}

func (g *RepeaterGui) GetModuleGui() widgets.QWidget_ITF {

	g.repeaterTabs = widgets.NewQTabWidget(nil)
	g.repeaterTabs.SetDocumentMode(true)
	g.repeaterTabs.ConnectCurrentChanged(g.changedTab)

	//g.repeaterTabs.AddTab(g.NewTab(), strconv.Itoa(g.tabNum))
	g.repeaterTabs.AddTab(g.NewEmptyTab(), "+")

	return g.repeaterTabs

}

func (g *RepeaterGui) changedTab(i int) {
	g.Sess.Debug("repeater", strconv.Itoa(i))
	if i == g.repeaterTabs.Count()-1 {
		// add a new tab before me
		g.repeaterTabs.InsertTab(g.repeaterTabs.Count()-1, g.NewEmptyTab(), strconv.Itoa(g.tabNum))
		g.tabNum = g.tabNum + 1
		g.repeaterTabs.SetCurrentIndex(g.repeaterTabs.Count() - 2)
	}
}

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

func (g *RepeaterGui) NewEmptyTab() widgets.QWidget_ITF {
	return g.NewTab("", "")

}

func (g *RepeaterGui) Name() string {
	return "Repeater"
}
