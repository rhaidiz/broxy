package model

import (
	"sync"

	bcore "github.com/rhaidiz/broxy/core"
	"github.com/therecipe/qt/core"
)

const (
	Type = int(core.Qt__UserRole) + 1<<iota
	Module
	Time
	Message
)

func (m *CustomTableModel) roleNames() map[int]*core.QByteArray {
	return map[int]*core.QByteArray{
		Type:    core.NewQByteArray2("Type", -1),
		Module:  core.NewQByteArray2("Module", -1),
		Time:    core.NewQByteArray2("Time", -1),
		Message: core.NewQByteArray2("Message", -1),
	}
}

type CustomTableModel struct {
	core.QAbstractTableModel

	modelData []bcore.Log

	_ func() `constructor:"init"`

	_ func(item bcore.Log) `signal:"addItem,auto"`
}

var mutex = &sync.Mutex{}

func (m *CustomTableModel) init() {
	m.modelData = []bcore.Log{}

	m.ConnectRoleNames(m.roleNames)
	m.ConnectRowCount(m.rowCount)
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
func (m *CustomTableModel) data(index *core.QModelIndex, role int) *core.QVariant {
	item := m.modelData[index.Row()]
	switch role {
	case Type:
		return core.NewQVariant14(item.Type)
	case Module:
		return core.NewQVariant14(item.ModuleName)
	case Time:
		return core.NewQVariant14(item.Time)
	case Message:
		return core.NewQVariant14(item.Message)
	}
	return core.NewQVariant()
}
