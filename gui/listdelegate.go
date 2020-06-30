package gui

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// Delegate represents the listView delegate to list projects
type Delegate struct {
	widgets.QStyledItemDelegate
	qApp *widgets.QApplication
}

// InitDelegate is used to initialize the delegate
func InitDelegate(qApp *widgets.QApplication) *Delegate {
	item := NewDelegate(nil) //will be generated in moc.go
	item.qApp = qApp
	item.ConnectPaint(item.paint)
	item.ConnectSizeHint(item.sizeHint)
	return item
}

func (i Delegate) paint(painter *gui.QPainter, option *widgets.QStyleOptionViewItem, index *core.QModelIndex) {

	//widgets.QStyleOptionViewItem options = option
	i.InitStyleOption(option, index)

	painter.Save()

	doc := gui.NewQTextDocument(nil)
	//doc.setHtml(options.text);
	//fontMetric := gui.NewQFontMetricsF(painter.Font())

	//localElidedText := fontMetric.ElidedText(option.Text(), core.Qt__ElideMiddle, 355, 0)
	doc.SetHtml(option.Text())

	/* Call this to get the focus rect and selection background. */
	option.SetText("")
	//option.GetWidget.Style().drawControl(widgets.QStyle__CE_ItemViewItem, option, painter)
	//w := widgets.NewQWidgetFromPointer(option.Pointer())
	//style := widgets.NewQStyle2()

	//style :=(*widgets.QApplication).Style();
	//w.Style().DrawControl(widgets.QStyle__CE_ItemViewItem, option, painter, nil)
	i.qApp.Style().DrawControl(widgets.QStyle__CE_ItemViewItem, option, painter, nil)

	// /* Draw using our rich text document. */
	painter.Translate3(float64(option.Rect().Left()), float64(option.Rect().Top()))
	// QRect clip(0, 0, options.rect.width(), options.rect.height());
	clip := core.NewQRectF4(0, 0, float64(option.Rect().Width()), float64(option.Rect().Height()))
	doc.DrawContents(painter, clip.QRectF_PTR())

	painter.Restore()
}

func (i Delegate) sizeHint(option *widgets.QStyleOptionViewItem, index *core.QModelIndex) *core.QSize {
	return core.NewQSize2(20,45);
}
