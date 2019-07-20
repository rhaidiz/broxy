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
					role: "Method"
					title: role
				}
				TableViewColumn {
					role: "Schema"
					title: role
				}
	
				TableViewColumn {
					role: "Path"
					title: role
				}
				TableViewColumn {
					role: "Status"
					title: role
				}
			}
	
		}
	}
}

