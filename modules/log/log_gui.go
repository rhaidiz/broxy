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

	Sess         *core.Session
	view         *quick.QQuickView
	logTableView *widgets.QTableView
}

func NewLogGui(s *core.Session) *LogGui {
	return &LogGui{
		Sess:         s,
		view:         quick.NewQQuickView(nil),
		logTableView: widgets.NewQTableView(nil),
	}
}

func (g *LogGui) SetTableModel(m *model.SortFilterModel) {
	g.logTableView.SetModel(m)
	g.logTableView.SetColumnWidth(model.Type, 50)
	g.logTableView.SetColumnWidth(model.Module, 100)
	g.logTableView.SetColumnWidth(model.Time, 150)
	g.logTableView.SetColumnWidth(model.Message, 200)
}

func (g *LogGui) GetModuleGui() widgets.QWidget_ITF {

	widget := widgets.NewQWidget(nil, 0)
	widget.SetLayout(widgets.NewQVBoxLayout())

	// table view written in qml
	g.view.SetTitle("Log table")
	g.view.SetResizeMode(quick.QQuickView__SizeRootObjectToView)
	g.logTableView.SetShowGrid(false)
	g.logTableView.VerticalHeader().Hide()
	g.logTableView.SetAlternatingRowColors(true)
	g.logTableView.SetEditTriggers(widgets.QAbstractItemView__NoEditTriggers)
	g.logTableView.SetSelectionBehavior(widgets.QAbstractItemView__SelectRows)
	g.logTableView.SetSelectionMode(widgets.QAbstractItemView__SingleSelection)
	g.logTableView.VerticalHeader().SetSectionResizeMode(widgets.QHeaderView__Fixed)
	g.logTableView.SetSortingEnabled(true)
	g.logTableView.VerticalHeader().SetDefaultSectionSize(10)
	g.logTableView.SortByColumn(model.Time, qtcore.Qt__DescendingOrder)
	//g.view.SetSource(qtcore.NewQUrl3("qrc:/qml/log.qml", 0))

	//widget.Layout().AddWidget(widgets.QWidget_CreateWindowContainer(g.view, nil, 0))
	widget.Layout().AddWidget(g.logTableView)

	return widget
}

func (m *LogGui) Name() string {
	return "Log"
}
