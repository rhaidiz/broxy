package coreproxy

import (
	"fmt"
	"os"
	"time"

	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/modules/coreproxy/model"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/quick"
	"github.com/therecipe/qt/widgets"
)

const (
	CopyURLLabel        = "Copy URL"
	CopyBaseURLLabel    = "Copy base URL"
	SendToRepeaterLabel = "Send to Repeater"
	ClearHistoryLabel   = "Clear History"
)

type CoreproxyGui struct {
	core.GuiModule

	_ func() `signal:"test,auto"`

	Sess *core.Session

	ControllerInit        func()
	StartProxy            func(bool)
	StopProxy             func()
	RowClicked            func(int)
	ApplyFilters          func(bool)
	ResetFilters          func(bool)
	CheckReqInterception  func(bool)
	CheckRespInterception func(bool)
	DownloadCAClicked     func(bool)
	RightItemClicked      func(string, int)
	settingsTab           *widgets.QTabWidget

	//_ func() `signal:"test,auto"`
	// history tab
	splitter               *widgets.QSplitter
	tableBridge            *TableBridge
	reqRespTab             *widgets.QTabWidget
	RequestTextEdit        *widgets.QPlainTextEdit
	EditedRequestTextEdit  *widgets.QPlainTextEdit
	ResponseTextEdit       *widgets.QPlainTextEdit
	EditedResponseTextEdit *widgets.QPlainTextEdit
	historyTab             *widgets.QTabWidget

	// Filter
	TextSearchLineEdit    *widgets.QLineEdit
	ApplyFiltersButton    *widgets.QPushButton
	ResetFiltersButton    *widgets.QPushButton
	S100CheckBox          *widgets.QCheckBox
	S200CheckBox          *widgets.QCheckBox
	S300CheckBox          *widgets.QCheckBox
	S400CheckBox          *widgets.QCheckBox
	S500CheckBox          *widgets.QCheckBox
	ShowExtensionLineEdit *widgets.QLineEdit
	HideExtensionLineEdit *widgets.QLineEdit
	ShowOnlyCheckBox      *widgets.QCheckBox
	HideOnlyCheckBox      *widgets.QCheckBox
	RightItemLabels       []string

	coreProxyGui *widgets.QTabWidget

	tableModel *model.CustomTableModel

	view *quick.QQuickView

	// settings tab
	ListenerLineEdit      *widgets.QLineEdit
	StartStopButton       *widgets.QPushButton
	ReqInterceptCheckBox  *widgets.QCheckBox
	RespInterceptCheckBox *widgets.QCheckBox
	DownloadCAButton      *widgets.QPushButton

	// interceptor
	ForwardButton           *widgets.QPushButton
	DropButton              *widgets.QPushButton
	InterceptorToggleButton *widgets.QPushButton
	InterceptorTextEdit     *widgets.QPlainTextEdit
	Toggle                  func(bool)
	Forward                 func(bool)
	Drop                    func(bool)
}

/*
 TableBridge is meant to expose QML signals from the tableview implemented in
 QML.  This is a very ugly workaround, but I want to use QML only for the table
 since it seems to perform better.
*/

type TableBridge struct {
	qtcore.QObject

	_ func(int)         `signal:"clicked,auto"`
	_ func(string, int) `signal:"rightItemClicked,auto"`

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

func (t *TableBridge) rightItemClicked(l string, r int) {
	t.coreGui.RightItemClicked(l, r)
}

func NewCoreproxyGui(s *core.Session) *CoreproxyGui {
	return &CoreproxyGui{
		RightItemLabels: []string{
			CopyURLLabel,
			CopyBaseURLLabel,
			ClearHistoryLabel,
			SendToRepeaterLabel,
		},
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

	g.ForwardButton = widgets.NewQPushButton2("Forward", nil)
	g.DropButton = widgets.NewQPushButton2("Drop", nil)
	g.InterceptorToggleButton = widgets.NewQPushButton2("Interceptor", nil)
	spacerItem := widgets.NewQSpacerItem(400, 20, widgets.QSizePolicy__Expanding, widgets.QSizePolicy__Minimum)

	hlayout.AddWidget(g.ForwardButton, 0, qtcore.Qt__AlignLeft)
	g.ForwardButton.ConnectClicked(g.Forward)
	hlayout.AddWidget(g.DropButton, 0, qtcore.Qt__AlignLeft)
	g.DropButton.ConnectClicked(g.Drop)
	hlayout.AddWidget(g.InterceptorToggleButton, 0, qtcore.Qt__AlignLeft)
	g.InterceptorToggleButton.ConnectClicked(g.Toggle)
	g.InterceptorToggleButton.SetAutoRepeat(true)
	g.InterceptorToggleButton.SetCheckable(true)
	hlayout.AddItem(spacerItem)

	vlayout.AddLayout(hlayout, 0)

	g.InterceptorTextEdit = widgets.NewQPlainTextEdit(nil)
	vlayout.AddWidget(g.InterceptorTextEdit, 0, 0)

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

	g.S100CheckBox = widgets.NewQCheckBox(nil)
	g.S100CheckBox.SetText("1xx")
	vlayout1.AddWidget(g.S100CheckBox, 0, qtcore.Qt__AlignLeft)

	g.S200CheckBox = widgets.NewQCheckBox(nil)
	g.S200CheckBox.SetText("2xx")
	vlayout1.AddWidget(g.S200CheckBox, 0, qtcore.Qt__AlignLeft)

	g.S300CheckBox = widgets.NewQCheckBox(nil)
	g.S300CheckBox.SetText("3xx")
	vlayout1.AddWidget(g.S300CheckBox, 0, qtcore.Qt__AlignLeft)

	g.S400CheckBox = widgets.NewQCheckBox(nil)
	g.S400CheckBox.SetText("4xx")
	vlayout1.AddWidget(g.S400CheckBox, 0, qtcore.Qt__AlignLeft)

	g.S500CheckBox = widgets.NewQCheckBox(nil)
	g.S500CheckBox.SetText("5xx")
	vlayout1.AddWidget(g.S500CheckBox, 0, qtcore.Qt__AlignLeft)

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
	g.ShowExtensionLineEdit = widgets.NewQLineEdit(nil)
	g.ShowExtensionLineEdit.SetMinimumSize(qtcore.NewQSize2(150, 0))
	g.ShowExtensionLineEdit.SetMaximumSize(qtcore.NewQSize2(150, 16777215))
	g.ShowExtensionLineEdit.SetBaseSize(qtcore.NewQSize2(0, 0))
	g.ShowExtensionLineEdit.SetText("")

	g.HideExtensionLineEdit = widgets.NewQLineEdit(nil)
	g.HideExtensionLineEdit.SetMinimumSize(qtcore.NewQSize2(150, 0))
	g.HideExtensionLineEdit.SetMaximumSize(qtcore.NewQSize2(150, 16777215))
	g.HideExtensionLineEdit.SetBaseSize(qtcore.NewQSize2(0, 0))
	g.HideExtensionLineEdit.SetText("")

	g.ShowOnlyCheckBox = widgets.NewQCheckBox(nil)
	g.ShowOnlyCheckBox.SetText("Show only")

	g.HideOnlyCheckBox = widgets.NewQCheckBox(nil)
	g.HideOnlyCheckBox.SetText("Hide")

	gridLayout.AddWidget(g.ShowExtensionLineEdit, 0, 1, 1)
	gridLayout.AddWidget(g.HideExtensionLineEdit, 1, 1, 1)

	gridLayout.AddWidget(g.ShowOnlyCheckBox, 0, 0, 1)
	gridLayout.AddWidget(g.HideOnlyCheckBox, 1, 0, 1)

	spacerItem := widgets.NewQSpacerItem(400, 20, widgets.QSizePolicy__Expanding, widgets.QSizePolicy__Minimum)
	gridLayout.AddItem(spacerItem, 0, 2, 1, 1, qtcore.Qt__AlignRight)

	vlayout1.AddLayout(gridLayout, 0)

	// Apply\Reset buttons
	g.ApplyFiltersButton = widgets.NewQPushButton2("Apply", nil)
	g.ApplyFiltersButton.ConnectClicked(g.ApplyFilters)

	gridLayout.AddWidget(g.ApplyFiltersButton, 2, 1, 1)

	g.ResetFiltersButton = widgets.NewQPushButton2("Reset", nil)
	g.ResetFiltersButton.ConnectClicked(g.ResetFilters)

	gridLayout.AddWidget(g.ResetFiltersButton, 2, 0, 1)

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

	g.StartStopButton = widgets.NewQPushButton2("Start", nil)
	g.StartStopButton.ConnectClicked(g.StartProxy)
	gridLayout.AddWidget(g.StartStopButton, 0, 1, 1)

	spacerItem := widgets.NewQSpacerItem(400, 20, widgets.QSizePolicy__Expanding, widgets.QSizePolicy__Minimum)
	gridLayout.AddItem(spacerItem, 0, 2, 1, 1, qtcore.Qt__AlignRight)

	vlayout1.AddLayout(gridLayout, 0)

	// interception settings
	label1 := widgets.NewQLabel(nil, 0)
	label1.SetFont(font)
	label1.SetText("Interception")
	vlayout1.AddWidget(label1, 0, qtcore.Qt__AlignLeft)

	g.ReqInterceptCheckBox = widgets.NewQCheckBox(nil)
	g.ReqInterceptCheckBox.SetText("Intercept requests")
	g.ReqInterceptCheckBox.ConnectClicked(g.CheckReqInterception)
	vlayout1.AddWidget(g.ReqInterceptCheckBox, 0, qtcore.Qt__AlignLeft)

	g.RespInterceptCheckBox = widgets.NewQCheckBox(nil)
	g.RespInterceptCheckBox.SetText("Intercept responses")
	g.RespInterceptCheckBox.ConnectClicked(g.CheckRespInterception)
	vlayout1.AddWidget(g.RespInterceptCheckBox, 0, qtcore.Qt__AlignLeft)

	label_ca := widgets.NewQLabel(nil, 0)
	label_ca.SetText("Certificate Authority")
	label_ca.SetFont(font)
	vlayout1.AddWidget(label_ca, 0, qtcore.Qt__AlignLeft)

	g.DownloadCAButton = widgets.NewQPushButton2("Download CA certificate", nil)
	g.DownloadCAButton.ConnectClicked(g.DownloadCAClicked)
	vlayout1.AddWidget(g.DownloadCAButton, 0, qtcore.Qt__AlignLeft)

	spacerItem1 := widgets.NewQSpacerItem(20, 40, widgets.QSizePolicy__Minimum, widgets.QSizePolicy__Expanding)
	vlayout1.AddItem(spacerItem1)

	return scrollArea
}

func (g *CoreproxyGui) SetRightClickMenu() {
	m := qtcore.NewQStringListModel2(g.RightItemLabels, nil)
	g.view.RootContext().SetContextProperty("MenuItems", m)
}

func (g *CoreproxyGui) SetTableModel(m *model.SortFilterModel) {
	//TODO: move the SetRightClickMenu somewhere that makes sense
	g.SetRightClickMenu()
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
	g.reqRespTab.AddTab(g.RequestTextEdit, "Request")
	g.RequestTextEdit.SetPlainText(req)
}

func (g *CoreproxyGui) ShowEditedReqTab(edited_req string) {
	g.reqRespTab.AddTab(g.EditedRequestTextEdit, "Edited Request")
	g.EditedRequestTextEdit.SetPlainText(edited_req)

}

func (g *CoreproxyGui) ShowRespTab(resp string) {
	g.reqRespTab.AddTab(g.ResponseTextEdit, "Response")
	g.ResponseTextEdit.SetPlainText(resp)
}

func (g *CoreproxyGui) ShowEditedRespTab(edited_resp string) {
	g.reqRespTab.AddTab(g.EditedResponseTextEdit, "Edited Response")
	g.EditedResponseTextEdit.SetPlainText(edited_resp)
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
	g.RequestTextEdit = widgets.NewQPlainTextEdit(nil)
	g.RequestTextEdit.SetReadOnly(true)
	g.ResponseTextEdit = widgets.NewQPlainTextEdit(nil)
	g.ResponseTextEdit.SetReadOnly(true)
	g.EditedRequestTextEdit = widgets.NewQPlainTextEdit(nil)
	g.EditedRequestTextEdit.SetReadOnly(true)
	g.EditedResponseTextEdit = widgets.NewQPlainTextEdit(nil)
	g.EditedResponseTextEdit.SetReadOnly(true)
	//g.reqRespTab.AddTab(g.RequestTextEdit, "Request")
	//g.reqRespTab.AddTab(g.EditedRequestTextEdit, "Edited Request")
	//g.reqRespTab.AddTab(g.ResponseTextEdit, "Response")

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

func (t *CoreproxyGui) FileSaveAs(s string) bool {
	var fileDialog = widgets.NewQFileDialog2(nil, "Save as...", "", "")
	fileDialog.SetAcceptMode(widgets.QFileDialog__AcceptSave)
	var mimeTypes = []string{"application/x-x509-ca-cert"}
	fileDialog.SetMimeTypeFilters(mimeTypes)
	fileDialog.SetDefaultSuffix("der")
	if fileDialog.Exec() != int(widgets.QDialog__Accepted) {
		return false
	}
	var fn = fileDialog.SelectedFiles()[0]

	f, err := os.Create(fn)
	if err != nil {
		return false
	}
	defer f.Close()
	_, err1 := f.WriteString(s)
	if err1 != nil {
		return false
	}
	return true
}

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
