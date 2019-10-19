import QtQuick 2.10				//Item
import QtQuick.Controls 1.4		//TableView
import QtQuick.Layouts 1.3		//ColumnLayout
import CustomQmlTypes 1.0		//CustomTableModel

Item {
	width: 600
	height: 500

	ColumnLayout {
		anchors.fill: parent

			SplitView {
	    	anchors.fill: parent
	      orientation: Qt.Vertical
	
			TableView {
				id: tableview
	
				Layout.fillWidth: true
				Layout.fillHeight: true
	
				sortIndicatorVisible: true
      	onSortIndicatorColumnChanged: tableview.model.sortTableView(tableview.getColumn(tableview.sortIndicatorColumn).role, sortIndicatorOrder)
				onSortIndicatorOrderChanged: tableview.model.sortTableView(tableview.getColumn(tableview.sortIndicatorColumn).role, sortIndicatorOrder)

	 			alternatingRowColors: false

				//onClicked: { tableBridge.clicked(currentRow) }
				Connections{
				  	target: tableview.selection
				  	onSelectionChanged: {tableBridge.clicked(tableview.currentRow) }
				}

				Keys.onUpPressed: {
					if(tableview.currentRow > 0)
			  	   tableview.currentRow--
						 selection.clear()
						 selection.select(tableview.currentRow)
				}
				 
				 Keys.onDownPressed: {
					if(tableview.currentRow < tableview.rowCount - 1)
				     tableview.currentRow++
						 selection.clear()
						 selection.select(tableview.currentRow)
				}
	
				model: MyModel
				
				TableViewColumn {
					role: "ID"
					title: role
					width: 40
				}
				TableViewColumn {
					role: "Host"
					title: role
				}
	
				TableViewColumn {
					role: "Method"
					title: role
					width: 80
				}
				TableViewColumn {
					role: "Path"
					title: role
				}
	
				TableViewColumn {
					role: "Params"
					title: role
					width: 60
				}
				TableViewColumn {
					role: "Edit"
					title: role
					width: 60
				}
				TableViewColumn {
					role: "Status"
					title: role
					width: 100
				}
				TableViewColumn {
					role: "Length"
					title: role
					width: 80
				}
			}
	
		}
	}
}

