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

	repeaterTabs   *widgets.QTabWidget
	goBtn          *widgets.QPushButton
	cancelBtn      *widgets.QPushButton
	HostLine       *widgets.QLineEdit
	RequestEditor  *widgets.QPlainTextEdit
	ResponseEditor *widgets.QPlainTextEdit
	tabNum         int

	_ func(i int) `signal:"changedTab"`

	GoClick func(bool)
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

func (g *RepeaterGui) AddNewTab(host string, request string) {
	g.tabNum = g.tabNum + 1
	g.repeaterTabs.InsertTab(g.repeaterTabs.Count()-1, g.NewTab(host, request), host)
	g.repeaterTabs.SetCurrentIndex(g.repeaterTabs.Count() - 2)
}

func (g *RepeaterGui) NewTab(host string, request string) widgets.QWidget_ITF {
	mainWidget := widgets.NewQWidget(nil, 0)
	vlayout := widgets.NewQVBoxLayout()
	vlayout.SetContentsMargins(11, 11, 11, 11)
	mainWidget.SetLayout(vlayout)

	hlayout := widgets.NewQHBoxLayout()

	g.goBtn = widgets.NewQPushButton2("Go", nil)
	g.cancelBtn = widgets.NewQPushButton2("Cancel", nil)
	g.goBtn.ConnectClicked(g.GoClick)
	hlayout.AddWidget(g.goBtn, 0, 0)
	hlayout.AddWidget(g.cancelBtn, 0, 0)

	g.HostLine = widgets.NewQLineEdit(nil)
	g.HostLine.SetText(host)
	hlayout.AddWidget(g.HostLine, 0, 0)

	vlayout.AddLayout(hlayout, 0)

	splitter := widgets.NewQSplitter(nil)
	splitter.SetOrientation(qtcore.Qt__Horizontal)

	g.RequestEditor = widgets.NewQPlainTextEdit(nil)
	g.RequestEditor.SetPlainText(request)
	g.ResponseEditor = widgets.NewQPlainTextEdit(nil)
	g.ResponseEditor.SetReadOnly(true)
	splitter.AddWidget(g.RequestEditor)
	splitter.AddWidget(g.ResponseEditor)

	vlayout.AddWidget(splitter, 0, 0)

	return mainWidget
}

func (g *RepeaterGui) NewEmptyTab() widgets.QWidget_ITF {
	return g.NewTab("", "")

}

func (g *RepeaterGui) Name() string {
	return "Repeater"
}
