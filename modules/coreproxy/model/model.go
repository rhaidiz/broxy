package model

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/therecipe/qt/core"
)

type Request struct {
	Url           *url.URL
	Proto         string
	Method        string
	Host          string
	Headers       http.Header
	ContentLength int64
	Body          []byte
	Extension     string
	Params        bool
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
	u1 := fmt.Sprintf("%v", r.Url)
	ret := fmt.Sprintf("%s %s %s\nHost: %s\n", r.Method, u1[len(r.Url.Scheme)+2:], r.Proto, r.Host)
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
	ID = iota
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

func (m *CustomTableModel) init() {
	m.modelData = []HttpItem{}
	m.hashMap = make(map[int64]*HttpItem)

	m.ConnectHeaderData(m.headerData)
	m.ConnectRowCount(m.rowCount)
	m.ConnectColumnCount(m.columnCount)
	m.ConnectData(m.data)
}

func (m *CustomTableModel) headerData(section int, orientation core.Qt__Orientation, role int) *core.QVariant {
	if role != int(core.Qt__DisplayRole) || orientation == core.Qt__Vertical {
		return m.HeaderDataDefault(section, orientation, role)
	}
	switch section {
	case ID:
		return core.NewQVariant1("ID")
	case Host:
		return core.NewQVariant1("Host")
	case Method:
		return core.NewQVariant1("Method")
	case Path:
		return core.NewQVariant1("Path")
	case Params:
		return core.NewQVariant1("Params")
	case Edit:
		return core.NewQVariant1("Edit")
	case Status:
		return core.NewQVariant1("Status")
	case Length:
		return core.NewQVariant1("Length")
	}
	return core.NewQVariant()
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
}

func (m *CustomTableModel) addItem(item *HttpItem, i int64) {
	mutex.Lock()
	defer mutex.Unlock()
	m.BeginInsertRows(core.NewQModelIndex(), len(m.modelData), len(m.modelData))
	m.hashMap[i] = item
	m.modelData = append(m.modelData, *item)
	m.EndInsertRows()
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
	return 8
}
func (m *CustomTableModel) data(index *core.QModelIndex, role int) *core.QVariant {
	if role == int(core.Qt__TextAlignmentRole) &&
		(index.Column() == Method ||
			index.Column() == Params ||
			index.Column() == Edit ||
			index.Column() == Length) {
		return core.NewQVariant1(int64(core.Qt__AlignCenter))
	}
	if role != int(core.Qt__DisplayRole) {
		return core.NewQVariant()
	}

	item := m.modelData[index.Row()]
	switch index.Column() {
	case ID:
		return core.NewQVariant1(item.ID)
	case Host:
		return core.NewQVariant1(item.Req.Host)
	case Method:
		return core.NewQVariant1(item.Req.Method)
	case Path:
		return core.NewQVariant1(item.Req.Url.Path)
	case Params:
		if item.Req.Params {
			return core.NewQVariant1("✓")
		}
		return core.NewQVariant1("")
	case Edit:
		if item.EditedReq != nil || item.EditedResp != nil {
			return core.NewQVariant1("✓")
		}
		return core.NewQVariant1("")
	case Status:
		if item.Resp != nil {
			return core.NewQVariant1(item.Resp.Status)
		}
	case Length:
		if item.Resp != nil {
			return core.NewQVariant1(fmt.Sprintf("%d", item.Resp.ContentLength))
		}
	}
	return core.NewQVariant()
}
