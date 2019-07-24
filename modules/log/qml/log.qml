import QtQuick 2.10				//Item
import QtQuick.Controls 1.4		//TableView
import QtQuick.Layouts 1.3		//ColumnLayout
import CustomQmlTypes 1.0		//CustomTableModel


			TableView {
				id: tableview
	
				Layout.fillWidth: true
				Layout.fillHeight: true
	
	 			alternatingRowColors: false

				sortIndicatorColumn: 2
				sortIndicatorOrder : Qt.DescendingOrder
				sortIndicatorVisible: true
      	onSortIndicatorColumnChanged: tableview.model.sortTableView(tableview.getColumn(tableview.sortIndicatorColumn).role, sortIndicatorOrder)
				onSortIndicatorOrderChanged: tableview.model.sortTableView(tableview.getColumn(tableview.sortIndicatorColumn).role, sortIndicatorOrder)

				model: MyModel
	
				TableViewColumn {
					role: "Type"
					title: role
				}
				TableViewColumn {
					role: "Module"
					title: role
				}
	
				TableViewColumn {
					role: "Time"
					title: role
				}
				TableViewColumn {
					role: "Message"
					title: role
				}
			}
