package model

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/therecipe/qt/core"
)

type Request struct {
	Proto         string
	Method        string
	Path          string
	Schema        string
	Host          string
	Headers       http.Header
	ContentLength int64
	Body          []byte
	Extension     string
}

func (r *Request) ToString() string {
	/*
		Metho Path Proto
		Host
		Headers

		Body
	*/
	if r == nil {
		return ""
	}
	path, err := url.Parse(r.Path)
	if err != nil {
		return "Url parser error"
	}
	ret := fmt.Sprintf("%s %s %s\nHost: %s\n", r.Method, path, r.Proto, r.Host)
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
	StatusCode    int
	Headers       http.Header
	ContentLength int64
	Body          []byte
}

func (r *Response) ToString() string {
	/*
		Proto Status
		Headers

		Body
	*/
	if r == nil {
		return ""
	}
	ret := fmt.Sprintf("%s %s\n", r.Proto, r.Status)
	for k, v := range r.Headers {
		values := ""
		for _, s := range v {
			values = values + s
		}
		ret = ret + fmt.Sprintf("%s: %s\n", k, values)
	}
	ret = ret + fmt.Sprintf("Content-Length: %d\n", r.ContentLength)
	if len(r.Body) > 0 {
		ret = ret + fmt.Sprintf("\n%s", string(r.Body))
	}
	return ret
}

type HttpItem struct {
	core.QObject
	ID         int
	Req        *Request
	Resp       *Response
	EditedReq  *Request
	EditedResp *Response
}

//func NewHttpItem2() *HttpItem {
// func (f *HttpItem) init() {
// 	empty_req := &Request{
// 		Proto:         "",
// 		Method:        "",
// 		Path:          "",
// 		Schema:        "",
// 		Host:          "",
// 		ContentLength: -1,
// 	}
// 	empty_resp := &Response{
// 		Proto:         "",
// 		Status:        "",
// 		StatusCode:    0,
// 		ContentLength: -1,
// 	}
// 	f.ID = 0
// 	f.Req = empty_req
// 	f.Resp = empty_resp
// 	f.EditedReq = empty_req
// 	f.EditedResp = empty_resp
// }

const (
	ID = int(core.Qt__UserRole) + 1<<iota
	Host
	Method
	Path
	Params
	Edit
	Status
	Length
)

func (m *CustomTableModel) row(i *HttpItem) int {
	for index, item := range m.modelData {
		if item.Pointer() == i.Pointer() {
			return index
		}
	}
	return 0
}

func (m *CustomTableModel) roleNames() map[int]*core.QByteArray {
	return map[int]*core.QByteArray{
		ID:     core.NewQByteArray2("ID", -1),
		Host:   core.NewQByteArray2("Host", -1),
		Method: core.NewQByteArray2("Method", -1),
		Path:   core.NewQByteArray2("Path", -1),
		Params: core.NewQByteArray2("Params", -1),
		Edit:   core.NewQByteArray2("Edit", -1),
		Status: core.NewQByteArray2("Status", -1),
		Length: core.NewQByteArray2("Length", -1),
		//LastName:  core.NewQByteArray2("LastName", -1),
	}
}

type CustomTableModel struct {
	core.QAbstractTableModel
	_ func() `constructor:"init"`

	modelData []HttpItem
	hashMap   map[int64]*HttpItem

	_ func(item *HttpItem, i int64) `signal:"addItem,auto"`
	_ func(item *HttpItem, i int64) `signal:editItem,auto"`
	_ func()                        `signal:clearHistory,auto"`
}

var mutex = &sync.Mutex{}

//func (m *CustomTableModel) GetIndex(i int) *HttpItem {
//	return &m.modelData[i]
//}

func init() {
	CustomTableModel_QmlRegisterType2("CustomQmlTypes", 1, 0, "CustomTableModel")
}

func (m *CustomTableModel) init() {
	m.modelData = []HttpItem{}
	m.hashMap = make(map[int64]*HttpItem)

	m.ConnectRoleNames(m.roleNames)
	//m.ConnectHeaderData(m.headerData)
	m.ConnectRowCount(m.rowCount)
	m.ConnectColumnCount(m.columnCount)
	m.ConnectData(m.data)
}

func (m *CustomTableModel) GetReqResp(i int) (*Request, *Request, *Response, *Response) {
	if i >= 0 {
		return m.modelData[i].Req, m.modelData[i].EditedReq, m.modelData[i].Resp, m.modelData[i].EditedResp
	}
	return nil, nil, nil, nil
}

func (m *CustomTableModel) AddReq(r *http.Request, i int64) {
	mutex.Lock()
	defer mutex.Unlock()

	// save the request and its body
	//m.hashMap[i] = HttpItem{Req: r, ReqBody: bodyBytes}
}

func (m *CustomTableModel) clearHistory() {
	mutex.Lock()
	defer mutex.Unlock()
	m.BeginRemoveRows(core.NewQModelIndex(), 0, len(m.modelData))
	m.modelData = []HttpItem{}
	m.hashMap = make(map[int64]*HttpItem)
	m.EndRemoveRows()
	fmt.Println("done %d", len(m.modelData))
}

func (m *CustomTableModel) addItem(item *HttpItem, i int64) {
	mutex.Lock()
	defer mutex.Unlock()
	m.BeginInsertRows(core.NewQModelIndex(), len(m.modelData), len(m.modelData))
	m.hashMap[i] = item
	m.modelData = append(m.modelData, *item)
	m.EndInsertRows()
	fmt.Println("add item ", len(m.modelData))
}

func (m *CustomTableModel) editItem(item *HttpItem, i int64) {
	mutex.Lock()
	defer mutex.Unlock()

	row := m.row(m.hashMap[i])

	m.hashMap[i].Resp = item.Resp
	m.hashMap[i].EditedResp = item.EditedResp

	m.modelData[row].Resp = item.Resp
	m.modelData[row].EditedResp = item.EditedResp

	m.DataChanged(m.Index(row, 2, core.NewQModelIndex()), m.Index(row, 2, core.NewQModelIndex()), []int{Edit, Status, Length})
}

func (m *CustomTableModel) rowCount(*core.QModelIndex) int {
	return len(m.modelData)
}

func (m *CustomTableModel) columnCount(*core.QModelIndex) int {
	return 5
}
func (m *CustomTableModel) data(index *core.QModelIndex, role int) *core.QVariant {
	item := m.modelData[index.Row()]
	switch role {
	case ID:
		return core.NewQVariant7(item.ID)
	case Host:
		return core.NewQVariant14(item.Req.Host)
	case Method:
		return core.NewQVariant14(item.Req.Method)
	case Path:
		return core.NewQVariant14(item.Req.Path)
	case Params:
		//TODO fix me
		if false {
			return core.NewQVariant14("✓")
		}
		return core.NewQVariant14("")
	case Edit:
		if item.EditedReq != nil || item.EditedResp != nil {
			return core.NewQVariant14("✓")
		}
		return core.NewQVariant14("")
	case Status:
		if item.Resp != nil {
			return core.NewQVariant14(item.Resp.Status)
		}
	case Length:
		if item.Resp != nil {
			return core.NewQVariant14(fmt.Sprintf("%d", item.Resp.ContentLength))
		}
	}
	return core.NewQVariant()
}
