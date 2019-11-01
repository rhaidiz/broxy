package model

import (
	"github.com/therecipe/qt/core"
	"strings"
)

type SortFilterModel struct {
	core.QSortFilterProxyModel

	Custom *CustomTableModel

	searchtext string

	_ func() `constructor:"init"`

	_ func(column string, order core.Qt__SortOrder) `signal:"sortTableView"`
}

func init() {
	CustomTableModel_QmlRegisterType2("CustomQmlTypes", 1, 0, "SortFilterModel")
}

func (m *SortFilterModel) init() {
	m.Custom = NewCustomTableModel(nil)

	m.SetSourceModel(m.Custom)
	//m.SetSortRole(Time)
	//m.Sort(0, core.Qt__DescendingOrder)

	m.ConnectFilterAcceptsRow(m.filterAcceptsRow)
	m.ConnectSortTableView(m.sortTableView)
}

func (m *SortFilterModel) SetFilter(s string) {
	m.searchtext = s
	m.InvalidateFilter()
}

func (m *SortFilterModel) ResetFilters() {
	m.searchtext = ""
	m.InvalidateFilter()
}

func (m *SortFilterModel) filterAcceptsRow(sourceRow int, sourceParent *core.QModelIndex) bool {
	req, _, _, _ := m.Custom.GetReqResp(sourceRow)
	if strings.Contains(req.Host, m.searchtext) {
		return true
	}
	return false
}

func (m *SortFilterModel) sortTableView(column string, order core.Qt__SortOrder) {
	for k, v := range m.Custom.RoleNames() {
		if v.ConstData() == column {
			m.SetSortRole(k)
			m.Sort(0, order)
		}
	}
}
