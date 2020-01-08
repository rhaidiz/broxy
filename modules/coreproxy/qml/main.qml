import QtQuick 2.12
import QtQuick.Controls 1.4
import QtQuick.Controls 2.12
import QtQuick.Layouts 1.12
import CustomQmlTypes 1.0

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

				 rowDelegate: 
				 Rectangle{
				 color: styleData.selected ? "#448" : "transparent"
								 
                 MouseArea {
                      id: mouseArea
                      acceptedButtons: Qt.RightButton
                      anchors.fill: parent
                      propagateComposedEvents: true
                       onClicked: {
				 							tableview.selection.clear();
				 							tableview.currentRow = styleData.row
    		 							tableview.selection.select(styleData.row);
                         mouse.accepted = false
											 menu.popup()
                       }
                  }
									Menu {
    							    id: menu
											 Instantiator {
    										   model: MenuItems
    										   MenuItem {
    										      text: model.display
															onTriggered: tableBridge.rightItemClicked(model.display, tableview.currentRow)
    										   }
    										onObjectAdded: menu.insertItem(index, object)
    										onObjectRemoved: menu.removeItem(object)
												}
    							}
				 }
				
				
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

