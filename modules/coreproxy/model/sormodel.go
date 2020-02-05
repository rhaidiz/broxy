package model

import (
	"strings"

	"github.com/therecipe/qt/core"
)

type SortFilterModel struct {
	core.QSortFilterProxyModel

	Custom *CustomTableModel

	filter *Filter

	_ func() `constructor:"init"`

	_ func(column string, order core.Qt__SortOrder) `signal:"sortTableView"`
}

func (m *SortFilterModel) init() {
	m.Custom = NewCustomTableModel(nil)

	m.SetSourceModel(m.Custom)
	//m.SetSortRole(Time)
	//m.Sort(0, core.Qt__DescendingOrder)

	m.ConnectFilterAcceptsRow(m.filterAcceptsRow)
	m.ConnectSortTableView(m.sortTableView)
}

func (m *SortFilterModel) SetFilter(f *Filter) {
	m.filter = f
	m.InvalidateFilter()
}

func (m *SortFilterModel) ResetFilters() {
	m.InvalidateFilter()
}

func (m *SortFilterModel) filterAcceptsRow(sourceRow int, sourceParent *core.QModelIndex) bool {
	if m.filter == nil {
		return true
	}
	req, edited_req, resp, edited_resp := m.Custom.GetReqResp(sourceRow)

	// extension
	if (len(m.filter.Hide_ext) > 0 && m.filter.Hide_ext[req.Extension] == true) ||
		(len(m.filter.Show_ext) > 0 && m.filter.Show_ext[req.Extension] == false) {
		return false
	}

	// response status
	next := false //IMP: make me pretier
	for _, i := range m.filter.StatusCode {
		if resp != nil && ((resp.StatusCode <= i+99 && resp.StatusCode >= i) || resp.StatusCode > 599) {
			next = true
			break
		}
	}
	if resp != nil && !next {
		return false
	}

	// text search filter
	txt := ""
	if req != nil {
		txt = req.ToString()
	}
	if edited_req != nil {
		txt = txt + edited_req.ToString()
	}
	if resp != nil {
		txt = txt + resp.ToString()
	}
	if edited_resp != nil {
		txt = txt + edited_resp.ToString()
	}

	if !strings.Contains(txt, m.filter.Search) {
		return false
	}

	return true
}

func (m *SortFilterModel) sortTableView(column string, order core.Qt__SortOrder) {
	for k, v := range m.Custom.RoleNames() {
		if v.ConstData() == column {
			m.SetSortRole(k)
			m.Sort(0, order)
		}
	}
}
