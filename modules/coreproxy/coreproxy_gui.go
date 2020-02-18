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
	CopyURLLabel      = "Copy URL"
	CopyBaseURLLabel  = "Copy base URL"
	RepeatLabel       = "Repeat"
	ClearHistoryLabel = "Clear History"
)

// Gui represents the GUI of the main intercept proxy
type Gui struct {
	core.GuiModule

	Sess *core.Session

	rightClickLabels      [4]string
	ControllerInit        func()
	StartProxy            func(bool)
	StopProxy             func()
	RowClicked            func(int)
	ApplyFilters          func(bool)
	ResetFilters          func(bool)
	CheckReqInterception  func(bool)
	CheckRespInterception func(bool)
	CheckIgnoreHTTPS      func(bool)
	SaveCAClicked         func(bool)
	RightItemClicked      func(string, int)
	settingsTab           *widgets.QTabWidget

	// history tab
	historyTableView       *widgets.QTableView
	contextMenu            *widgets.QMenu
	splitter               *widgets.QSplitter
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

	coreProxyGui *widgets.QTabWidget

	tableModel *model.CustomTableModel

	view *quick.QQuickView

	// settings tab
	ListenerLineEdit      *widgets.QLineEdit
	StartStopButton       *widgets.QPushButton
	ReqInterceptCheckBox  *widgets.QCheckBox
	RespInterceptCheckBox *widgets.QCheckBox
	SaveCAButton          *widgets.QPushButton
	ignoreHTTPSCheckBox   *widgets.QCheckBox

	// interceptor
	ForwardButton           *widgets.QPushButton
	DropButton              *widgets.QPushButton
	InterceptorToggleButton *widgets.QPushButton
	InterceptorTextEdit     *widgets.QPlainTextEdit
	Toggle                  func(bool)
	Forward                 func(bool)
	Drop                    func(bool)
}

// NewGui creates a new Gui for the main intercetp proxy
func NewGui(s *core.Session) *Gui {
	return &Gui{
		Sess:             s,
		historyTableView: widgets.NewQTableView(nil),
		view:             quick.NewQQuickView(nil),
		rightClickLabels: [4]string{CopyURLLabel, CopyBaseURLLabel, RepeatLabel, ClearHistoryLabel},
	}
}

func (g *Gui) interceptorTabGui() widgets.QWidget_ITF {
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

func (g *Gui) filtersTabGui() widgets.QWidget_ITF {
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

	gridLayout.AddWidget(g.ShowExtensionLineEdit)
	gridLayout.AddWidget(g.HideExtensionLineEdit)

	gridLayout.AddWidget(g.ShowOnlyCheckBox)
	gridLayout.AddWidget(g.HideOnlyCheckBox)

	spacerItem := widgets.NewQSpacerItem(400, 20, widgets.QSizePolicy__Expanding, widgets.QSizePolicy__Minimum)
	gridLayout.AddItem(spacerItem, 0, 2, 1, 1, qtcore.Qt__AlignRight)

	vlayout1.AddLayout(gridLayout, 0)

	// Apply\Reset buttons
	g.ApplyFiltersButton = widgets.NewQPushButton2("Apply", nil)
	g.ApplyFiltersButton.ConnectClicked(g.ApplyFilters)

	gridLayout.AddWidget(g.ApplyFiltersButton)

	g.ResetFiltersButton = widgets.NewQPushButton2("Reset", nil)
	g.ResetFiltersButton.ConnectClicked(g.ResetFilters)

	gridLayout.AddWidget(g.ResetFiltersButton)

	spacerItem1 := widgets.NewQSpacerItem(20, 1000, widgets.QSizePolicy__Minimum, widgets.QSizePolicy__Expanding)
	vlayout1.AddItem(spacerItem1)

	return scrollArea
}

func (g *Gui) settingsTabGui() widgets.QWidget_ITF {
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

	label2 := widgets.NewQLabel(nil, 0)
	label2.SetObjectName("label2")
	label2.SetText("Description goes here")
	vlayout1.AddWidget(label2, 0, qtcore.Qt__AlignLeft)

	gridLayout := widgets.NewQGridLayout2()
	g.ListenerLineEdit = widgets.NewQLineEdit(nil)
	g.ListenerLineEdit.SetMinimumSize(qtcore.NewQSize2(150, 0))
	g.ListenerLineEdit.SetMaximumSize(qtcore.NewQSize2(150, 16777215))
	g.ListenerLineEdit.SetBaseSize(qtcore.NewQSize2(0, 0))
	g.ListenerLineEdit.SetText("127.0.0.1:8080")
	gridLayout.AddWidget(g.ListenerLineEdit)

	g.StartStopButton = widgets.NewQPushButton2("Start", nil)
	g.StartStopButton.ConnectClicked(g.StartProxy)
	gridLayout.AddWidget(g.StartStopButton)

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

	labelCA := widgets.NewQLabel(nil, 0)
	labelCA.SetText("Certificate Authority")
	labelCA.SetFont(font)
	vlayout1.AddWidget(labelCA, 0, qtcore.Qt__AlignLeft)

	g.SaveCAButton = widgets.NewQPushButton2("Save CA certificate", nil)
	g.SaveCAButton.ConnectClicked(g.SaveCAClicked)
	vlayout1.AddWidget(g.SaveCAButton, 0, qtcore.Qt__AlignLeft)

	g.ignoreHTTPSCheckBox = widgets.NewQCheckBox(nil)
	g.ignoreHTTPSCheckBox.SetText("Do not intercept HTTPS")
	g.ignoreHTTPSCheckBox.ConnectClicked(g.CheckIgnoreHTTPS)
	vlayout1.AddWidget(g.ignoreHTTPSCheckBox, 0, qtcore.Qt__AlignLeft)

	spacerItem1 := widgets.NewQSpacerItem(20, 40, widgets.QSizePolicy__Minimum, widgets.QSizePolicy__Expanding)
	vlayout1.AddItem(spacerItem1)

	return scrollArea
}

// SetRightClickMenu sets the menu items when right clicking an item in the history table
func (g *Gui) SetRightClickMenu() {
}

// SetTableModel sets the table model along with some column width to use in the history table
func (g *Gui) SetTableModel(m *model.SortFilterModel) {
	g.historyTableView.SetModel(m)
	g.historyTableView.SetColumnWidth(model.ID, 40)
	g.historyTableView.SetColumnWidth(model.Host, 200)
	g.historyTableView.SetColumnWidth(model.Method, 80)
	g.historyTableView.SetColumnWidth(model.Path, 200)
	g.historyTableView.SetColumnWidth(model.Params, 60)
	g.historyTableView.SetColumnWidth(model.Edit, 60)
	g.historyTableView.SetColumnWidth(model.Status, 80)
	g.historyTableView.SetColumnWidth(model.Length, 80)
	//TODO: move the SetRightClickMenu somewhere that makes sense
	g.SetRightClickMenu()
}

// HideAllTabs hides the tabs used to view details of a single row in the history table
func (g *Gui) HideAllTabs() {
	for i := g.reqRespTab.Count(); i != 0; i-- {
		g.reqRespTab.RemoveTab(i)
	}
}

// ShowReqTab shows the request tab for the currently selected item in the history table
func (g *Gui) ShowReqTab(req string) {
	g.reqRespTab.AddTab(g.RequestTextEdit, "Request")
	g.RequestTextEdit.SetPlainText(req)
}

// ShowEditedReqTab shows the edited request tab for the currently selected item in the history table
func (g *Gui) ShowEditedReqTab(editedReq string) {
	g.reqRespTab.AddTab(g.EditedRequestTextEdit, "Edited Request")
	g.EditedRequestTextEdit.SetPlainText(editedReq)

}

// ShowRespTab shows the response tab for the currently selected item in the history table
func (g *Gui) ShowRespTab(resp string) {
	g.reqRespTab.AddTab(g.ResponseTextEdit, "Response")
	g.ResponseTextEdit.SetPlainText(resp)
}

// ShowEditedRespTab shows the edited response tab for the currently selected item in the history table
func (g *Gui) ShowEditedRespTab(editedResp string) {
	g.reqRespTab.AddTab(g.EditedResponseTextEdit, "Edited Response")
	g.EditedResponseTextEdit.SetPlainText(editedResp)
}

func (g *Gui) customContextMenuRequested(p *qtcore.QPoint) {
	if g.contextMenu == nil {
		g.contextMenu = widgets.NewQMenu(nil)
		copyURLAction := g.contextMenu.AddAction(CopyURLLabel)
		copyURLAction.ConnectTriggered(func(b bool) {
			if len(g.historyTableView.SelectedIndexes()) > 0 {
				g.RightItemClicked(CopyURLLabel, g.historyTableView.SelectedIndexes()[0].Row())
			}
		})

		copyBaseURLAction := g.contextMenu.AddAction(CopyBaseURLLabel)
		copyBaseURLAction.ConnectTriggered(func(b bool) {
			if len(g.historyTableView.SelectedIndexes()) > 0 {
				g.RightItemClicked(CopyBaseURLLabel, g.historyTableView.SelectedIndexes()[0].Row())
			}
		})

		repeatAction := g.contextMenu.AddAction(RepeatLabel)
		repeatAction.ConnectTriggered(func(b bool) {
			if len(g.historyTableView.SelectedIndexes()) > 0 {
				g.RightItemClicked(RepeatLabel, g.historyTableView.SelectedIndexes()[0].Row())
			}
		})

		clearHistoryAction := g.contextMenu.AddAction(ClearHistoryLabel)
		clearHistoryAction.ConnectTriggered(func(b bool) {
			if len(g.historyTableView.SelectedIndexes()) > 0 {
				g.RightItemClicked(ClearHistoryLabel, g.historyTableView.SelectedIndexes()[0].Row())
			}
		})

	}
	p.SetY(p.Ry() + 15)
	g.contextMenu.Exec2(g.historyTableView.MapToGlobal(p), nil)
}

// GetModuleGui returns the Gui for the current module
func (g *Gui) GetModuleGui() widgets.QWidget_ITF {
	g.coreProxyGui = widgets.NewQTabWidget(nil)
	g.coreProxyGui.SetDocumentMode(true)

	g.historyTableView.SetShowGrid(false)
	g.historyTableView.VerticalHeader().Hide()
	g.historyTableView.SetAlternatingRowColors(true)
	g.historyTableView.ConnectClicked(func(index *qtcore.QModelIndex) {
		g.RowClicked(index.Row())
	})
	//g.historyTableView.ConnectActivated(model.ShowMessage2)
	g.historyTableView.ConnectCurrentChanged(func(current *qtcore.QModelIndex, prev *qtcore.QModelIndex) {
		g.historyTableView.ScrollTo(current, 0)
		g.RowClicked(current.Row())
	})
	g.historyTableView.SetEditTriggers(widgets.QAbstractItemView__NoEditTriggers)
	g.historyTableView.SetSelectionBehavior(widgets.QAbstractItemView__SelectRows)
	g.historyTableView.SetSelectionMode(widgets.QAbstractItemView__SingleSelection)
	g.historyTableView.SetContextMenuPolicy(qtcore.Qt__CustomContextMenu)
	g.historyTableView.ConnectCustomContextMenuRequested(g.customContextMenuRequested)
	g.historyTableView.SetSortingEnabled(true)
	g.historyTableView.VerticalHeader().SetSectionResizeMode(widgets.QHeaderView__Fixed)
	g.historyTableView.VerticalHeader().SetDefaultSectionSize(10)

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
	g.splitter.AddWidget(g.historyTableView)
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

	g.coreProxyGui.AddTab(g.interceptorTabGui(), "Interceptor")
	g.coreProxyGui.AddTab(g.historyTab, "History")
	g.coreProxyGui.AddTab(g.settingsTabGui(), "Settings")

	//IMP: make me pretier
	g.ControllerInit()

	return g.coreProxyGui
}

// FileSaveAs saves the CA file
func (g *Gui) FileSaveAs(s string) bool {
	var fileDialog = widgets.NewQFileDialog2(nil, "Save as...", "broxyca.pem", "PEM (*.pem)")
	fileDialog.SetAcceptMode(widgets.QFileDialog__AcceptSave)
	var mimeTypes = []string{"application/x-pem-file"}
	fileDialog.SetMimeTypeFilters(mimeTypes)
	fileDialog.SetDefaultSuffix("pem")
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

func (g *Gui) bench(b bool) {
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

// Title returns the time of this Gui
func (g *Gui) Title() string {
	return "Proxy"
}
