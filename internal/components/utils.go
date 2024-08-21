package components

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func updateTable(tableView *tview.Table, table *Table) {
	tableView.SetTitle(fmt.Sprintf("[blue]%s[[grey]%d[blue]]", table.Title, len(table.Rows)))

	tableView.Clear()
	cols, rows := len(table.Columns), len(table.Rows)
	for c := 0; c < cols; c++ {
		color := tcell.ColorWhite
		tableView.SetCell(0, c,
			tview.NewTableCell(table.Columns[c]).
				SetTextColor(color).
				SetStyle(tcell.StyleDefault.Foreground(tcell.ColorGrey).Bold(true)).
				SetAlign(tview.AlignLeft).SetExpansion(1).SetSelectable(false))
	}

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			color := tcell.ColorAqua
			tableView.SetCell(r+1, c,
				tview.NewTableCell(table.Rows[r][c]).
					SetTextColor(color).
					SetAlign(tview.AlignLeft))
		}
	}
}
