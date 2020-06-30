package model

import (
	"sync"

	bcore "github.com/rhaidiz/broxy/core"
	"github.com/therecipe/qt/core"
)

const (
	Type = iota
	Module
	Time
	Message
)

// CustomTableModel represents a table model used to populate the log QtTableView
type CustomTableModel struct {
	core.QAbstractTableModel

	modelData []bcore.Log

	_ func() `constructor:"init"`

	_ func(item bcore.Log) `signal:"addItem,auto"`
}

var mutex = &sync.Mutex{}

func (m *CustomTableModel) init() {
	m.modelData = []bcore.Log{}

	m.ConnectRowCount(m.rowCount)
	m.ConnectHeaderData(m.headerData)
	m.ConnectColumnCount(m.columnCount)
	m.ConnectData(m.data)
}

func (m *CustomTableModel) addItem(item bcore.Log) {
	mutex.Lock()
	defer mutex.Unlock()
	m.BeginInsertRows(core.NewQModelIndex(), len(m.modelData), len(m.modelData))
	m.modelData = append(m.modelData, item)
	m.EndInsertRows()
}

func (m *CustomTableModel) rowCount(*core.QModelIndex) int {
	return len(m.modelData)
}

func (m *CustomTableModel) columnCount(*core.QModelIndex) int {
	return 4
}

func (m *CustomTableModel) headerData(section int, orientation core.Qt__Orientation, role int) *core.QVariant {
	if role != int(core.Qt__DisplayRole) || orientation == core.Qt__Vertical {
		return m.HeaderDataDefault(section, orientation, role)
	}
	switch section {
	case Type:
		return core.NewQVariant1("Type")
	case Module:
		return core.NewQVariant1("Module")
	case Time:
		return core.NewQVariant1("Time")
	case Message:
		return core.NewQVariant1("Message")
	}
	return core.NewQVariant()
}

func (m *CustomTableModel) data(index *core.QModelIndex, role int) *core.QVariant {
	if role == int(core.Qt__TextAlignmentRole) &&
		(index.Column() == Type) {
		return core.NewQVariant1(int64(core.Qt__AlignCenter))
	}
	if role != int(core.Qt__DisplayRole) {
		return core.NewQVariant()
	}
	item := m.modelData[index.Row()]
	switch index.Column() {
	case Type:
		return core.NewQVariant1(item.Type)
	case Module:
		return core.NewQVariant1(item.ModuleName)
	case Time:
		return core.NewQVariant1(item.Time)
	case Message:
		return core.NewQVariant1(item.Message)
	}
	return core.NewQVariant()
}
