package log

import (
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/log/model"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/quick"
	"github.com/therecipe/qt/widgets"
)

// Gui represents the Gui of the log module
type Gui struct {
	core.GuiModule

	Sess         *core.Session
	view         *quick.QQuickView
	logTableView *widgets.QTableView
}

// NewGui returns a Gui of the main log module
func NewGui(s *core.Session) *Gui {
	return &Gui{
		Sess:         s,
		view:         quick.NewQQuickView(nil),
		logTableView: widgets.NewQTableView(nil),
	}
}

func (g *Gui) GetSettings() interface{} {
	return nil
}
// SetTableModel sets the table model along with some column width to use in the history table
func (g *Gui) SetTableModel(m *model.SortFilterModel) {
	g.logTableView.SetModel(m)
	g.logTableView.SetColumnWidth(model.Type, 50)
	g.logTableView.SetColumnWidth(model.Module, 100)
	g.logTableView.SetColumnWidth(model.Time, 150)
	g.logTableView.SetColumnWidth(model.Message, 200)
}

// GetModuleGui returns the Gui for the current module
func (g *Gui) GetModuleGui() interface{}  {

	widget := widgets.NewQWidget(nil, 0)
	widget.SetLayout(widgets.NewQVBoxLayout())

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

	widget.Layout().AddWidget(g.logTableView)

	return widget
}

// Title returns the time of this Gui
func (g *Gui) Title() string {
	return "Log"
}
