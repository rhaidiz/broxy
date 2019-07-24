package log

import (
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/log/model"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/quick"
	"github.com/therecipe/qt/widgets"
)

type LogGui struct {
	core.GuiModule

	Sess *core.Session
	view *quick.QQuickView
}

func NewLogGui(s *core.Session) *LogGui {
	return &LogGui{
		Sess: s,
		view: quick.NewQQuickView(nil),
	}
}

func (g *LogGui) SetTableModel(m *model.SortFilterModel) {
	g.view.RootContext().SetContextProperty("MyModel", m)
}

func (g *LogGui) GetModuleGui() widgets.QWidget_ITF {

	widget := widgets.NewQWidget(nil, 0)
	widget.SetLayout(widgets.NewQVBoxLayout())

	// table view written in qml
	g.view.SetTitle("Log table")
	g.view.SetResizeMode(quick.QQuickView__SizeRootObjectToView)
	g.view.SetSource(qtcore.NewQUrl3("qrc:/qml/log.qml", 0))

	widget.Layout().AddWidget(widgets.QWidget_CreateWindowContainer(g.view, nil, 0))

	return widget
}

func (m *LogGui) Name() string {
	return "Log"
}
