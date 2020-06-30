package model

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/therecipe/qt/core"
)

var History []*Row

// used to map id to actual rows in the table
var HashMapHistory  = make(map[int64]int)

type Row struct {
	ID				int64
	Req        		*Request
	Resp       		*Response
	EditedReq  		*Request
	EditedResp 		*Response
}

// Request represents an HTTP request logged in the history
type Request struct {
	ID			  int64
	URL           *url.URL
	Proto         string
	Method        string
	Host          string
	Headers       http.Header
	ContentLength int64
	Body          []byte
	Extension     string
	Params        bool
	IP						string
}

// ToString returns a string representation of an HTTP request logged in the history
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
	//u1 := fmt.Sprintf("%v", r.URL)
	ret := fmt.Sprintf("%s %s %s\nHost: %s\n", r.Method, r.URL.Path, r.Proto, r.Host)
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

// Response represents an HTTP response logged in the history
type Response struct {
	ID			  int64
	Proto         string
	Status        string
	StatusCode    int
	Headers       http.Header
	ContentLength int64
	Body          []byte
}

// ToString returns a string representation of an HTTP response logged in the history
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

// HTTPItem represent an item in the history table
type HTTPItem struct {
	core.QObject
	ID         int
	Req        *Request
	Resp       *Response
	EditedReq  *Request
	EditedResp *Response
}

const (
	ID = iota
	Host
	Method
	Path
	Params
	Edit
	Status
	Length
	Ip
)

/*func (m *CustomTableModel) row(i *HTTPItem) int {
	for index, item := range m.modelData {
		if item.Pointer() == i.Pointer() {
			return index
		}
	}
	return 0
}*/

// CustomTableModel represents a table model used to populate the history QtTableView
type CustomTableModel struct {
	core.QAbstractTableModel
	_ func() `constructor:"init"`

	//modelData []HTTPItem
	//hashMap   map[int64]*HTTPItem

/*	_ func(item *HTTPItem, i int64) `signal:"addItem,auto"`
	_ func(item *HTTPItem, i int64) `signal:editItem,auto"`*/
	_ func()                        `signal:clearHistory,auto"`
}

var mutex = &sync.Mutex{}

func (m *CustomTableModel) init() {
	//m.modelData = []HTTPItem{}
	//m.hashMap = make(map[int64]*HTTPItem)
	//HashMapHistory = make(map[int64]int)

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
	case Ip:
		return core.NewQVariant1("IP")
	}
	return core.NewQVariant()
}

// GetReqResp retursn request, response, edited request and edited response for a given context.ID
func (m *CustomTableModel) GetReqResp(i int64) (*Request, *Request, *Response, *Response) {
	if val, ok := HashMapHistory[i]; ok {
		return History[val].Req, History[val].EditedReq, History[val].Resp, History[val].EditedResp
	}else{
		for row, item := range History{
			//fmt.Printf("ID: %d\n", item.ID)
			//fmt.Printf("i: %d\n", i)
			if item.ID == i{
				HashMapHistory[i] = row
				return History[row].Req, History[row].EditedReq, History[row].Resp, History[row].EditedResp
			}
		}
	}
	//fmt.Println("here")
	return nil, nil, nil, nil
}

func (m *CustomTableModel) clearHistory() {
	mutex.Lock()
	defer mutex.Unlock()
	/*m.BeginRemoveRows(core.NewQModelIndex(), 0, len(m.modelData))
	m.modelData = []HTTPItem{}
	m.hashMap = make(map[int64]*HTTPItem)
	m.EndRemoveRows()*/
}

// AddRequest adds a Request to the history
func (m *CustomTableModel) AddRequest(r *Request, id int64){
	mutex.Lock()
	defer mutex.Unlock()
	m.BeginInsertRows(core.NewQModelIndex(), len(History), len(History))
	// create a new row with the request
	t := &Row{ID:id, Req:r}
	History = append(History, t)
	// save the i to a mapping with hasmap
	HashMapHistory[id] = len(History)-1
	m.EndInsertRows()
}

// AddEditedRequest adds an edited Request to the history
func (m *CustomTableModel) AddEditedRequest(r *Request, id int64){
	mutex.Lock()
	defer mutex.Unlock()
	row := HashMapHistory[id]
	History[row].EditedReq = r
	m.DataChanged(m.Index(row, 2, core.NewQModelIndex()), m.Index(row, 2, core.NewQModelIndex()), []int{Edit, Status, Length})
}

// AddEditedResponse adds an edited Response to the history
func (m *CustomTableModel) AddEditedResponse(r *Response, id int64){
	mutex.Lock()
	defer mutex.Unlock()
	row := HashMapHistory[id]
	History[row].EditedResp = r
	m.DataChanged(m.Index(row, 2, core.NewQModelIndex()), m.Index(row, 2, core.NewQModelIndex()), []int{Edit, Status, Length})
}

// AddResponse adds a Response to the history
func (m *CustomTableModel) AddResponse(r *Response, id int64){
	mutex.Lock()
	defer mutex.Unlock()
	row := HashMapHistory[id]
	History[row].Resp = r
	m.DataChanged(m.Index(row, 2, core.NewQModelIndex()), m.Index(row, 2, core.NewQModelIndex()), []int{Edit, Status, Length})
}

/*func (m *CustomTableModel) addItem(item *HTTPItem, i int64) {
	mutex.Lock()
	defer mutex.Unlock()
	m.BeginInsertRows(core.NewQModelIndex(), len(m.modelData), len(m.modelData))
	m.hashMap[i] = item
	m.modelData = append(m.modelData, *item)
	m.EndInsertRows()
}

func (m *CustomTableModel) editItem(item *HTTPItem, i int64) {
	mutex.Lock()
	defer mutex.Unlock()

	row := m.row(m.hashMap[i])

	m.hashMap[i].Resp = item.Resp
	m.hashMap[i].EditedResp = item.EditedResp

	m.modelData[row].Resp = item.Resp
	m.modelData[row].EditedResp = item.EditedResp

	m.DataChanged(m.Index(row, 2, core.NewQModelIndex()), m.Index(row, 2, core.NewQModelIndex()), []int{Edit, Status, Length})
}*/

func (m *CustomTableModel) rowCount(*core.QModelIndex) int {
	return len(History)
}

func (m *CustomTableModel) columnCount(*core.QModelIndex) int {
	return 9
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

	item := History[index.Row()]
	switch index.Column() {
	case ID:
		return core.NewQVariant1(item.ID)
	case Host:
		return core.NewQVariant1(fmt.Sprintf("%s://%s",item.Req.URL.Scheme, item.Req.Host))
	case Method:
		return core.NewQVariant1(item.Req.Method)
	case Path:
		return core.NewQVariant1(item.Req.URL.Path)
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
	case Ip:
		return core.NewQVariant1(item.Req.IP)
	}
	return core.NewQVariant()
}
