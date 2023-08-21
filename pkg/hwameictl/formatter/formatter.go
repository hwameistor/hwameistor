package formatter

import (
	"os"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
)

// ParameterTableLineLength the length for each line about parameter table
const ParameterTableLineLength = 4

func buildDefaultTable() table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	//t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true

	return t
}

func PrintParameters(title string, parameters map[string]interface{}) {
	t := buildDefaultTable()

	if title != "" {
		t.SetTitle(title)
	}

	// Sort the keys
	keys := make([]string, 0, len(parameters))
	for key := range parameters {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Get the count of the rows
	length := len(parameters) / ParameterTableLineLength
	if len(parameters)%ParameterTableLineLength != 0 {
		length++
	}
	rows := make([]table.Row, length)

	// Parse map values to the table rows
	rowIndex, rowLength := 0, 0
	for _, key := range keys {
		if rowLength == ParameterTableLineLength {
			// set length to 0, and switch to next row
			rowIndex, rowLength = rowIndex+1, 0
		}
		rows[rowIndex], rowLength = append(rows[rowIndex], key, parameters[key]), rowLength+1
		if rowLength != ParameterTableLineLength {
			rows[rowIndex] = append(rows[rowIndex], " ")
		}
	}

	t.AppendRows(rows)
	t.Render()
}

func PrintTable(title string, header table.Row, rows []table.Row) {
	t := buildDefaultTable()

	if title != "" {
		t.SetTitle(title)
	}

	t.AppendHeader(header)
	t.AppendRows(rows)
	t.Render()
}
