package formatter

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
)

func buildDefaultTable() table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	//t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true

	return t
}

func PrintTable(title string, header table.Row, rows []table.Row) {
	t := buildDefaultTable()

	t.SetTitle(title)
	t.AppendHeader(header)
	t.AppendRows(rows)

	t.Render()
}
