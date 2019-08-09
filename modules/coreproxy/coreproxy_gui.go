package coreproxy

import (
	"fmt"
	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/coreproxy/model"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/quick"
	"github.com/therecipe/qt/widgets"
	"time"
)

type CoreproxyGui struct {
	core.GuiModule

	_ func() `signal:"test,auto"`

	Sess *core.Session

	StartProxy func(bool)
	StopProxy  func()
	RowClicked func(int)

	historyTab  *widgets.QTabWidget
	settingsTab *widgets.QTabWidget

	//_ func() `signal:"test,auto"`
	// history tab
	splitter     *widgets.QSplitter
	historyTable *widgets.QTreeView
	tableBridge  *TableBridge
	reqRespTab   *widgets.QTabWidget
	RequestText  *widgets.QPlainTextEdit
	ResponseText *widgets.QPlainTextEdit

	coreProxyGui *widgets.QTabWidget

	//widget      *widgets.QWidget
	//tableview   *widgets.QTableView
	//buttonStart *widgets.QPushButton
	//buttonStop  *widgets.QPushButton

	tableModel *model.CustomTableModel

	view *quick.QQuickView

	// settings tab
	ListenerLineEdit *widgets.QLineEdit
	StartStopBtn     *widgets.QPushButton

	// interceptor
	ForwardBtn        *widgets.QPushButton
	DropBtn           *widgets.QPushButton
	InterceptorToggle *widgets.QPushButton
	InterceptorEditor *widgets.QPlainTextEdit
	Toggle            func(bool)
	Forward           func(bool)
	Drop              func(bool)
}

/*
 TableBridge is meant to expose QML signals from the tableview implemented in
 QML.  This is a very ugly workaround, but I want to use QML only for the table
 since it seems to perform better.
*/

type TableBridge struct {
	qtcore.QObject

	_ func(int) `signal:"clicked,auto"`

	coreGui *CoreproxyGui
}

func (t *TableBridge) setParent(p *CoreproxyGui) {
	t.coreGui = p
}

func (t *TableBridge) clicked(r int) {
	if t.coreGui != nil {
		t.coreGui.RowClicked(r)
	}
}

func NewCoreproxyGui(s *core.Session) *CoreproxyGui {
	return &CoreproxyGui{
		Sess:        s,
		tableBridge: NewTableBridge(nil),
		view:        quick.NewQQuickView(nil),
	}
}

func (g *CoreproxyGui) intercetorTabGui() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	vlayout := widgets.NewQVBoxLayout()
	widget.SetLayout(vlayout)

	widget.SetContentsMargins(0, 0, 0, 0)

	hlayout := widgets.NewQHBoxLayout()

	g.ForwardBtn = widgets.NewQPushButton2("Forward", nil)
	g.DropBtn = widgets.NewQPushButton2("Drop", nil)
	g.InterceptorToggle = widgets.NewQPushButton2("Interceptor", nil)
	spacerItem := widgets.NewQSpacerItem(400, 20, widgets.QSizePolicy__Expanding, widgets.QSizePolicy__Minimum)

	hlayout.AddWidget(g.ForwardBtn, 0, qtcore.Qt__AlignLeft)
	g.ForwardBtn.ConnectClicked(g.Forward)
	hlayout.AddWidget(g.DropBtn, 0, qtcore.Qt__AlignLeft)
	g.DropBtn.ConnectClicked(g.Drop)
	hlayout.AddWidget(g.InterceptorToggle, 0, qtcore.Qt__AlignLeft)
	g.InterceptorToggle.ConnectClicked(g.Toggle)
	g.InterceptorToggle.SetAutoRepeat(true)
	g.InterceptorToggle.SetCheckable(true)
	hlayout.AddItem(spacerItem)

	vlayout.AddLayout(hlayout, 0)

	g.InterceptorEditor = widgets.NewQPlainTextEdit(nil)
	vlayout.AddWidget(g.InterceptorEditor, 0, 0)

	return widget

}

func (g *CoreproxyGui) settingsTabGui() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	vlayout1 := widgets.NewQVBoxLayout()
	widget.SetLayout(vlayout1)

	widget.SetContentsMargins(11, 11, 11, 11)
	////widget.SetSpacing(6)
	widget.SetObjectName("verticalLayout")

	label := widgets.NewQLabel(nil, 0)
	font := gui.NewQFont()
	font.SetPointSize(20)
	font.SetBold(true)
	font.SetWeight(75)
	label.SetFont(font)
	label.SetObjectName("label")
	label.SetText("Proxy Listener")
	vlayout1.AddWidget(label, 0, qtcore.Qt__AlignLeft)

	label_2 := widgets.NewQLabel(nil, 0)
	label_2.SetObjectName("label_2")
	label_2.SetText("Description goes here")
	vlayout1.AddWidget(label_2, 0, qtcore.Qt__AlignLeft)

	gridLayout := widgets.NewQGridLayout2()
	g.ListenerLineEdit = widgets.NewQLineEdit(nil)
	g.ListenerLineEdit.SetMinimumSize(qtcore.NewQSize2(150, 0))
	g.ListenerLineEdit.SetMaximumSize(qtcore.NewQSize2(150, 16777215))
	g.ListenerLineEdit.SetBaseSize(qtcore.NewQSize2(0, 0))
	g.ListenerLineEdit.SetText("127.0.0.1:8080")
	gridLayout.AddWidget(g.ListenerLineEdit, 0, 0, 1)

	g.StartStopBtn = widgets.NewQPushButton2("Start", nil)
	g.StartStopBtn.ConnectClicked(g.StartProxy)
	gridLayout.AddWidget(g.StartStopBtn, 0, 1, 1)

	spacerItem := widgets.NewQSpacerItem(400, 20, widgets.QSizePolicy__Expanding, widgets.QSizePolicy__Minimum)
	gridLayout.AddItem(spacerItem, 0, 2, 1, 1, qtcore.Qt__AlignRight)

	vlayout1.AddLayout(gridLayout, 0)

	spacerItem1 := widgets.NewQSpacerItem(20, 40, widgets.QSizePolicy__Minimum, widgets.QSizePolicy__Expanding)
	vlayout1.AddItem(spacerItem1)

	return widget
}

// func (g *CoreproxyGui) SetTableModel2(m *model.CustomTableModel) {
// 	g.tableModel = m
// }

func (g *CoreproxyGui) SetTableModel(m *model.SortFilterModel) {
	g.view.RootContext().SetContextProperty("MyModel", m)
	g.tableBridge.setParent(g)
	g.view.RootContext().SetContextProperty("tableBridge", g.tableBridge)
}

func (g *CoreproxyGui) GetModuleGui() widgets.QWidget_ITF {
	g.coreProxyGui = widgets.NewQTabWidget(nil)
	g.coreProxyGui.SetDocumentMode(true)

	// table view written in qml
	g.view.SetTitle("tableview Example")
	g.view.SetResizeMode(quick.QQuickView__SizeRootObjectToView)
	g.view.SetSource(qtcore.NewQUrl3("qrc:/qml/main.qml", 0))

	// request\response tabs with text editor
	g.reqRespTab = widgets.NewQTabWidget(nil)
	g.reqRespTab.SetDocumentMode(true)
	g.RequestText = widgets.NewQPlainTextEdit(nil)
	g.RequestText.SetReadOnly(true)
	g.ResponseText = widgets.NewQPlainTextEdit(nil)
	g.ResponseText.SetReadOnly(true)
	g.reqRespTab.AddTab(g.RequestText, "Request")
	g.reqRespTab.AddTab(g.ResponseText, "Response")

	// the splitter for tab history
	g.splitter = widgets.NewQSplitter(nil)
	g.splitter.SetOrientation(qtcore.Qt__Vertical)
	g.splitter.AddWidget(widgets.QWidget_CreateWindowContainer(g.view, nil, 0))
	g.splitter.AddWidget(g.reqRespTab)
	var sizes []int
	sizes = make([]int, 2)
	sizes[0] = 1 * g.splitter.SizeHint().Height()
	sizes[1] = 1 * g.splitter.SizeHint().Height()
	g.splitter.SetSizes(sizes)

	// a start button
	// g.startBtn = widgets.NewQPushButton2("Start", nil)
	// g.startBtn.ConnectClicked(g.StartProxy)

	g.coreProxyGui.AddTab(g.intercetorTabGui(), "Interceptor")
	g.coreProxyGui.AddTab(g.splitter, "History")
	g.coreProxyGui.AddTab(g.settingsTabGui(), "Settings")

	return g.coreProxyGui
}

//func (g *CoreproxyGui) GetModuleGui2() widgets.QWidget_ITF {
//
//	g.coreProxyGui = widgets.NewQTabWidget(nil)
//	g.coreProxyGui.SetDocumentMode(true)
//
//	g.historyTable = widgets.NewQTreeView(nil)
//	g.reqRespTab = widgets.NewQTabWidget(nil)
//	g.reqRespTab.SetDocumentMode(true)
//
//	g.historyTable.SetModel(g.tableModel)
//	//g.historyTable.VerticalHeader().Hide()
//	//g.historyTable.SetSelectionBehavior(widgets.QAbstractItemView__SelectRows)
//	//g.historyTable.SetVerticalScrollMode(widgets.QAbstractItemView__ScrollPerPixel)
//	//g.historyTable.SetSelectionMode(widgets.QAbstractItemView__SingleSelection)
//	//g.historyTable.ConnectSelectionChanged(func(s *qtcore.QItemSelection, ds *qtcore.QItemSelection) {
//	//	fmt.Println("sel changed")
//	//	//g.requestText.SetPlainText(g.tableModel.GetIndex(index.Row()).Path)
//	//})
//
//	//g.requestText = widgets.NewQPlainTextEdit(nil)
//	//g.responseText = widgets.NewQPlainTextEdit(nil)
//
//	//g.reqRespTab.AddTab(g.requestText, "Request")
//	//g.reqRespTab.AddTab(g.responseText, "Response")
//
//	g.splitter = widgets.NewQSplitter(nil)
//
//	g.splitter.SetOrientation(qtcore.Qt__Vertical)
//
//	g.splitter.AddWidget(g.historyTable)
//	g.splitter.AddWidget(g.reqRespTab)
//	var sizes []int
//	sizes = make([]int, 2)
//	sizes[0] = 1 * g.splitter.SizeHint().Height()
//	sizes[1] = 1 * g.splitter.SizeHint().Height()
//	g.splitter.SetSizes(sizes)
//
//	g.startBtn = widgets.NewQPushButton2("Start", nil)
//	g.startBtn.ConnectClicked(g.StartProxy)
//
//	bench := widgets.NewQPushButton2("bench", nil)
//	//bench.ConnectClicked(g.blabla)
//
//	g.coreProxyGui.AddTab(g.splitter, "History")
//	g.coreProxyGui.AddTab(g.startBtn, "Settings")
//	g.coreProxyGui.AddTab(bench, "Bench")
//
//	return g.coreProxyGui
//}

func (g *CoreproxyGui) Name() string {
	return "Proxy"
}

func (g *CoreproxyGui) bench(b bool) {
	fmt.Println("start here")

	s := time.Now()
	rows := g.tableModel.RowCount(qtcore.NewQModelIndex())
	columns := g.tableModel.ColumnCount(qtcore.NewQModelIndex())

	for i := 0; i < rows; i++ {

		for j := 0; j < columns; j++ {

			value := g.tableModel.Data(g.tableModel.Index(i, j, qtcore.NewQModelIndex()), 0)
			fmt.Println(value.ToString())

		}

	}

	elapsed := time.Since(s)
	fmt.Println(elapsed)

}

//func (g *CoreproxyGui) blabla(b bool) {
//	for i := 0; i < 10; i++ {
//		time.Sleep(100 * time.Millisecond)
//	}
//}
