package model

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/therecipe/qt/core"
)

type Request struct {
	Proto         string
	Method        string
	Path          string
	Schema        string
	Host          string
	Headers       map[string][]string
	ContentLength int64
	Body          []byte
}

func (r *Request) ToString() string {
	/*
		Metho Path Proto
		Host
		Headers

		Body
	*/
	ret := fmt.Sprintf("%s %s %s\nHost: %s\n", r.Method, r.Path, r.Proto, r.Host)
	for k, v := range r.Headers {
		values := ""
		for _, s := range v {
			values = values + s
		}
		ret = ret + fmt.Sprintf("%s: %s\n", k, values)
	}
	if len(r.Body) > 0 {
		ret = ret + fmt.Sprintf("\n%s", string(r.Body))
	}
	return ret
}

type Response struct {
	Proto         string
	Status        string
	Headers       map[string][]string
	ContentLength int64
	Body          []byte
}

func (r *Response) ToString() string {
	/*
		Proto Status
		Headers

		Body
	*/
	ret := fmt.Sprintf("%s %s\n", r.Proto, r.Status)
	for k, v := range r.Headers {
		values := ""
		for _, s := range v {
			values = values + s
		}
		ret = ret + fmt.Sprintf("%s: %s\n", k, values)
	}
	ret = ret + fmt.Sprintf("Content-Length: %s\n", r.ContentLength)
	if len(r.Body) > 0 {
		ret = ret + fmt.Sprintf("\n%s", string(r.Body))
	}
	return ret
}

type HItem struct {
	core.QObject
	Req  *Request
	Resp *Response
}

const (
	Method = int(core.Qt__UserRole) + 1<<iota
	Path
	Schema
	Status
)

func (m *CustomTableModel) row(i *HItem) int {
	for index, item := range m.modelData {
		if item.Pointer() == i.Pointer() {
			return index
		}
	}
	return 0
}

func (m *CustomTableModel) roleNames() map[int]*core.QByteArray {
	return map[int]*core.QByteArray{
		Method: core.NewQByteArray2("Method", -1),
		Path:   core.NewQByteArray2("Path", -1),
		Schema: core.NewQByteArray2("Schema", -1),
		Status: core.NewQByteArray2("Status", -1),
		//LastName:  core.NewQByteArray2("LastName", -1),
	}
}

type CustomTableModel struct {
	core.QAbstractTableModel
	_ func() `constructor:"init"`

	modelData []HItem

	_ func()                     `signal:"add,auto"`
	_ func(item *HItem, i int64) `signal:"addItem,auto"`
	_ func(item *HItem, i int64) `signal:editItem,auto"`
	_ func(item *HItem, i int64) `signal:prova,auto"`

	hashMap map[int64]*HItem

	_ func(row int) `signal:"test,auto"`
}

var mutex = &sync.Mutex{}

//func (m *CustomTableModel) GetIndex(i int) *HItem {
//	return &m.modelData[i]
//}

func init() {
	CustomTableModel_QmlRegisterType2("CustomQmlTypes", 1, 0, "CustomTableModel")
}

func (m *CustomTableModel) init() {
	m.modelData = []HItem{}
	m.hashMap = make(map[int64]*HItem)

	m.ConnectRoleNames(m.roleNames)
	//m.ConnectHeaderData(m.headerData)
	m.ConnectRowCount(m.rowCount)
	m.ConnectColumnCount(m.columnCount)
	m.ConnectData(m.data)
}

func (m *CustomTableModel) GetReqResp(i int) (*Request, *Response) {
	return m.modelData[i].Req, m.modelData[i].Resp
}

func (m *CustomTableModel) AddReq(r *http.Request, i int64) {
	mutex.Lock()
	defer mutex.Unlock()

	// save the request and its body
	//m.hashMap[i] = HItem{Req: r, ReqBody: bodyBytes}
}

//func (m *CustomTableModel) AddResp(resp *http.Response)
//{
//
//}

func (m *CustomTableModel) add() {
	m.BeginInsertRows(core.NewQModelIndex(), len(m.modelData), len(m.modelData))
	m.modelData = append(m.modelData, HItem{Req: &Request{Method: "blabla123"}})
	m.EndInsertRows()
}

func (m *CustomTableModel) addItem(item *HItem, i int64) {
	mutex.Lock()
	defer mutex.Unlock()
	m.BeginInsertRows(core.NewQModelIndex(), len(m.modelData), len(m.modelData))
	m.hashMap[i] = item
	m.modelData = append(m.modelData, *item)
	m.EndInsertRows()
}

func (m *CustomTableModel) test(row int) {
	// this updates to add the request when it arrives
	if len(m.modelData) == 0 {
		return
	}

	m.DataChanged(m.Index(row, 3, core.NewQModelIndex()), m.Index(row, 3, core.NewQModelIndex()), []int{Method, Path, Schema, Status})
	//m.DataChanged(m.Index(len(m.modelData)-1, 0, core.NewQModelIndex()), m.Index(len(m.modelData)-1, 1, core.NewQModelIndex()), []int{Method, Path, Schema, Status})

}

func (m *CustomTableModel) prova(item *HItem, i int64) {
	fmt.Println(item.Resp)
}

func (m *CustomTableModel) editItem(item *HItem, i int64) {
	mutex.Lock()
	defer mutex.Unlock()

	row := m.row(m.hashMap[i])

	m.hashMap[i].Resp = item.Resp
	m.modelData[row].Resp = item.Resp
	m.DataChanged(m.Index(row, 3, core.NewQModelIndex()), m.Index(row, 3, core.NewQModelIndex()), []int{Method, Path, Schema, Status})
}

func (m *CustomTableModel) headerData(section int, orientation core.Qt__Orientation, role int) *core.QVariant {
	if role != int(core.Qt__DisplayRole) || orientation == core.Qt__Vertical {
		return m.HeaderDataDefault(section, orientation, role)
	}

	switch section {
	case 0:
		return core.NewQVariant14("Method")
	case 1:
		return core.NewQVariant14("Schema")
	case 2:
		return core.NewQVariant14("Path")
	case 3:
		return core.NewQVariant14("Status")
	}
	return core.NewQVariant()
}

func (m *CustomTableModel) rowCount(*core.QModelIndex) int {
	return len(m.modelData)
}

func (m *CustomTableModel) columnCount(*core.QModelIndex) int {
	return 4
}
func (m *CustomTableModel) data(index *core.QModelIndex, role int) *core.QVariant {
	item := m.modelData[index.Row()]
	switch role {
	case Method:
		return core.NewQVariant14(item.Req.Method)
	case Path:
		return core.NewQVariant14(item.Req.Path)
	case Schema:
		return core.NewQVariant14(item.Req.Schema)
	case Status:
		if item.Resp != nil {
			return core.NewQVariant14(item.Resp.Status)
		}
	}
	return core.NewQVariant()
}

func (m *CustomTableModel) data2(index *core.QModelIndex, role int) *core.QVariant {
	if role != int(core.Qt__DisplayRole) {
		return core.NewQVariant()
	}

	item := m.modelData[index.Row()]
	switch m.HeaderData(index.Column(), core.Qt__Horizontal, role).ToString() {
	case "Method":
		return core.NewQVariant14(item.Req.Method)
	case "Schema":
		return core.NewQVariant14("test2")
	case "Path":
		return core.NewQVariant14("test3")
	case "Status":
		if item.Resp != nil {
			return core.NewQVariant14(item.Resp.Status)
		}
		return core.NewQVariant14("")
	}
	return core.NewQVariant()
}
