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

	ControllerInit func()
	StartProxy     func(bool)
	StopProxy      func()
	RowClicked     func(int)
	ApplyFilters   func(bool)
	ResetFilters   func(bool)

	settingsTab *widgets.QTabWidget

	//_ func() `signal:"test,auto"`
	// history tab
	splitter           *widgets.QSplitter
	historyTable       *widgets.QTreeView
	tableBridge        *TableBridge
	reqRespTab         *widgets.QTabWidget
	RequestText        *widgets.QPlainTextEdit
	EditedRequestText  *widgets.QPlainTextEdit
	ResponseText       *widgets.QPlainTextEdit
	EditedResponseText *widgets.QPlainTextEdit
	historyTab         *widgets.QTabWidget

	// Filter
	TextSearchLineEdit      *widgets.QLineEdit
	ApplyFiltersBtn         *widgets.QPushButton
	ResetFiltersBtn         *widgets.QPushButton
	Checkbox_status_100     *widgets.QCheckBox
	Checkbox_status_200     *widgets.QCheckBox
	Checkbox_status_300     *widgets.QCheckBox
	Checkbox_status_400     *widgets.QCheckBox
	Checkbox_status_500     *widgets.QCheckBox
	LineEdit_show_extension *widgets.QLineEdit
	LineEdit_hide_extension *widgets.QLineEdit
	Checkbox_show_only      *widgets.QCheckBox
	Checkbox_hide_only      *widgets.QCheckBox

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

func (g *CoreproxyGui) filtersTabGui() widgets.QWidget_ITF {
	scrollArea := widgets.NewQScrollArea(nil)
	scrollArea.SetWidgetResizable(true)
	scrollArea.SetGeometry2(10, 10, 200, 200)
	scrollAreaWidget := widgets.NewQWidget(nil, 0)
	vlayout1 := widgets.NewQVBoxLayout()
	scrollAreaWidget.SetLayout(vlayout1)
	scrollArea.SetWidget(scrollAreaWidget)

	//TODO: implement scope
	// in-scope checkbox
	// label := widgets.NewQLabel(nil, 0)
	// font := gui.NewQFont()
	// font.SetPointSize(20)
	// font.SetBold(true)
	// font.SetWeight(75)
	// label.SetFont(font)
	// label.SetText("Req type")
	// vlayout1.AddWidget(label, 0, qtcore.Qt__AlignLeft)

	// Checkbox_only_scope := widgets.NewQCheckBox(nil)
	// Checkbox_only_scope.SetText("Only in-scope")
	// vlayout1.AddWidget(Checkbox_only_scope, 0, qtcore.Qt__AlignLeft)

	// search text
	label2 := widgets.NewQLabel(nil, 0)
	font2 := gui.NewQFont()
	font2.SetPointSize(20)
	font2.SetBold(true)
	font2.SetWeight(75)
	label2.SetFont(font2)
	label2.SetText("Text search")
	vlayout1.AddWidget(label2, 0, qtcore.Qt__AlignLeft)

	g.TextSearchLineEdit = widgets.NewQLineEdit(nil)
	g.TextSearchLineEdit.SetMinimumSize(qtcore.NewQSize2(150, 0))
	g.TextSearchLineEdit.SetMaximumSize(qtcore.NewQSize2(150, 16777215))
	g.TextSearchLineEdit.SetBaseSize(qtcore.NewQSize2(0, 0))
	g.TextSearchLineEdit.SetText("")
	vlayout1.AddWidget(g.TextSearchLineEdit, 0, qtcore.Qt__AlignLeft)

	// Status
	label3 := widgets.NewQLabel(nil, 0)
	font3 := gui.NewQFont()
	font3.SetPointSize(20)
	font3.SetBold(true)
	font3.SetWeight(75)
	label3.SetFont(font3)
	label3.SetText("Status")
	vlayout1.AddWidget(label3, 0, qtcore.Qt__AlignLeft)

	g.Checkbox_status_100 = widgets.NewQCheckBox(nil)
	g.Checkbox_status_100.SetText("1xx")
	vlayout1.AddWidget(g.Checkbox_status_100, 0, qtcore.Qt__AlignLeft)

	g.Checkbox_status_200 = widgets.NewQCheckBox(nil)
	g.Checkbox_status_200.SetText("2xx")
	vlayout1.AddWidget(g.Checkbox_status_200, 0, qtcore.Qt__AlignLeft)

	g.Checkbox_status_300 = widgets.NewQCheckBox(nil)
	g.Checkbox_status_300.SetText("3xx")
	vlayout1.AddWidget(g.Checkbox_status_300, 0, qtcore.Qt__AlignLeft)

	g.Checkbox_status_400 = widgets.NewQCheckBox(nil)
	g.Checkbox_status_400.SetText("4xx")
	vlayout1.AddWidget(g.Checkbox_status_400, 0, qtcore.Qt__AlignLeft)

	g.Checkbox_status_500 = widgets.NewQCheckBox(nil)
	g.Checkbox_status_500.SetText("5xx")
	vlayout1.AddWidget(g.Checkbox_status_500, 0, qtcore.Qt__AlignLeft)

	// Extensions
	label4 := widgets.NewQLabel(nil, 0)
	font4 := gui.NewQFont()
	font4.SetPointSize(20)
	font4.SetBold(true)
	font4.SetWeight(75)
	label4.SetFont(font4)
	label4.SetText("Extension")

	vlayout1.AddWidget(label4, 0, qtcore.Qt__AlignLeft)

	gridLayout := widgets.NewQGridLayout2()
	g.LineEdit_show_extension = widgets.NewQLineEdit(nil)
	g.LineEdit_show_extension.SetMinimumSize(qtcore.NewQSize2(150, 0))
	g.LineEdit_show_extension.SetMaximumSize(qtcore.NewQSize2(150, 16777215))
	g.LineEdit_show_extension.SetBaseSize(qtcore.NewQSize2(0, 0))
	g.LineEdit_show_extension.SetText("")

	g.LineEdit_hide_extension = widgets.NewQLineEdit(nil)
	g.LineEdit_hide_extension.SetMinimumSize(qtcore.NewQSize2(150, 0))
	g.LineEdit_hide_extension.SetMaximumSize(qtcore.NewQSize2(150, 16777215))
	g.LineEdit_hide_extension.SetBaseSize(qtcore.NewQSize2(0, 0))
	g.LineEdit_hide_extension.SetText("")

	g.Checkbox_show_only = widgets.NewQCheckBox(nil)
	g.Checkbox_show_only.SetText("Show only")

	g.Checkbox_hide_only = widgets.NewQCheckBox(nil)
	g.Checkbox_hide_only.SetText("Hide")

	gridLayout.AddWidget(g.LineEdit_show_extension, 0, 1, 1)
	gridLayout.AddWidget(g.LineEdit_hide_extension, 1, 1, 1)

	gridLayout.AddWidget(g.Checkbox_show_only, 0, 0, 1)
	gridLayout.AddWidget(g.Checkbox_hide_only, 1, 0, 1)

	spacerItem := widgets.NewQSpacerItem(400, 20, widgets.QSizePolicy__Expanding, widgets.QSizePolicy__Minimum)
	gridLayout.AddItem(spacerItem, 0, 2, 1, 1, qtcore.Qt__AlignRight)

	vlayout1.AddLayout(gridLayout, 0)

	// Apply\Reset buttons
	g.ApplyFiltersBtn = widgets.NewQPushButton2("Apply", nil)
	g.ApplyFiltersBtn.ConnectClicked(g.ApplyFilters)

	gridLayout.AddWidget(g.ApplyFiltersBtn, 2, 1, 1)

	g.ResetFiltersBtn = widgets.NewQPushButton2("Reset", nil)
	g.ResetFiltersBtn.ConnectClicked(g.ResetFilters)

	gridLayout.AddWidget(g.ResetFiltersBtn, 2, 0, 1)

	spacerItem1 := widgets.NewQSpacerItem(20, 1000, widgets.QSizePolicy__Minimum, widgets.QSizePolicy__Expanding)
	vlayout1.AddItem(spacerItem1)

	return scrollArea
}

func (g *CoreproxyGui) settingsTabGui() widgets.QWidget_ITF {
	scrollArea := widgets.NewQScrollArea(nil)
	scrollArea.SetWidgetResizable(true)
	scrollArea.SetGeometry2(10, 10, 200, 200)
	scrollAreaWidget := widgets.NewQWidget(nil, 0)
	vlayout1 := widgets.NewQVBoxLayout()
	scrollAreaWidget.SetLayout(vlayout1)
	scrollArea.SetWidget(scrollAreaWidget)

	scrollArea.SetContentsMargins(11, 11, 11, 11)
	////widget.SetSpacing(6)
	scrollArea.SetObjectName("verticalLayout")

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

	return scrollArea
}

func (g *CoreproxyGui) SetTableModel(m *model.SortFilterModel) {
	g.view.RootContext().SetContextProperty("MyModel", m)
	g.tableBridge.setParent(g)
	g.view.RootContext().SetContextProperty("tableBridge", g.tableBridge)
}

func (g *CoreproxyGui) HideAllTabs() {
	for i := g.reqRespTab.Count(); i != 0; i-- {
		g.reqRespTab.RemoveTab(i)
	}
}

func (g *CoreproxyGui) ShowReqTab(req string) {
	g.reqRespTab.AddTab(g.RequestText, "Request")
	g.RequestText.SetPlainText(req)
}

func (g *CoreproxyGui) ShowEditedReqTab(edited_req string) {
	g.reqRespTab.AddTab(g.EditedRequestText, "Edited Request")
	g.EditedRequestText.SetPlainText(edited_req)

}

func (g *CoreproxyGui) ShowRespTab(resp string) {
	g.reqRespTab.AddTab(g.ResponseText, "Response")
	g.ResponseText.SetPlainText(resp)
}

func (g *CoreproxyGui) ShowEditedRespTab(edited_resp string) {
	g.reqRespTab.AddTab(g.EditedResponseText, "Edited Response")
	g.EditedResponseText.SetPlainText(edited_resp)
}

func (g *CoreproxyGui) GetModuleGui() widgets.QWidget_ITF {
	g.coreProxyGui = widgets.NewQTabWidget(nil)
	g.coreProxyGui.SetDocumentMode(true)

	// table view written in qml
	g.view.SetTitle("tableview Example")
	g.view.SetResizeMode(quick.QQuickView__SizeRootObjectToView)
	g.view.SetSource(qtcore.NewQUrl3("qrc:/qml/main.qml", 0))

	// history tab with filters
	g.historyTab = widgets.NewQTabWidget(nil)
	g.historyTab.SetDocumentMode(true)

	// request\response tabs with text editor
	g.reqRespTab = widgets.NewQTabWidget(nil)
	g.reqRespTab.SetDocumentMode(true)
	g.RequestText = widgets.NewQPlainTextEdit(nil)
	g.RequestText.SetReadOnly(true)
	g.ResponseText = widgets.NewQPlainTextEdit(nil)
	g.ResponseText.SetReadOnly(true)
	g.EditedRequestText = widgets.NewQPlainTextEdit(nil)
	g.EditedRequestText.SetReadOnly(true)
	g.EditedResponseText = widgets.NewQPlainTextEdit(nil)
	g.EditedResponseText.SetReadOnly(true)
	//g.reqRespTab.AddTab(g.RequestText, "Request")
	//g.reqRespTab.AddTab(g.EditedRequestText, "Edited Request")
	//g.reqRespTab.AddTab(g.ResponseText, "Response")

	// the splitter for tab history
	g.splitter = widgets.NewQSplitter(nil)
	g.splitter.SetOrientation(qtcore.Qt__Vertical)
	g.splitter.AddWidget(widgets.QWidget_CreateWindowContainer(g.view, nil, 0))
	g.splitter.AddWidget(g.reqRespTab)

	g.historyTab.AddTab(g.splitter, "History")
	g.historyTab.AddTab(g.filtersTabGui(), "Filters")

	var sizes []int
	sizes = make([]int, 2)
	sizes[0] = 1 * g.splitter.SizeHint().Height()
	sizes[1] = 1 * g.splitter.SizeHint().Height()
	g.splitter.SetSizes(sizes)

	// a start button
	// g.startBtn = widgets.NewQPushButton2("Start", nil)
	// g.startBtn.ConnectClicked(g.StartProxy)

	g.coreProxyGui.AddTab(g.intercetorTabGui(), "Interceptor")
	g.coreProxyGui.AddTab(g.historyTab, "History")
	g.coreProxyGui.AddTab(g.settingsTabGui(), "Settings")

	//IMP: make me pretier
	g.ControllerInit()

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
