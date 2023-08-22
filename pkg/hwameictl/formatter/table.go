package formatter

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"os"
)

// ParameterTableLineLength the length for each line about parameter table
const ParameterTableLineLength = 4

type Parameter struct {
	Key   interface{}
	Value interface{}
}

func buildDefaultTable() table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Set the header's format
	t.Style().Format.Header = text.FormatDefault
	return t
}

func PrintParameters(title string, parameters []Parameter) {
	t := buildDefaultTable()

	if title != "" {
		t.SetTitle(title)
	}

	// Get the length of the rows
	length := len(parameters) / ParameterTableLineLength
	if len(parameters)%ParameterTableLineLength != 0 {
		length++
	}
	rows := make([]table.Row, length)

	// Parse parameter values to the table rows
	rowIndex, rowLength := 0, 0
	for _, parameter := range parameters {
		if rowLength == ParameterTableLineLength {
			// set length to 0, and switch to next row
			rowIndex, rowLength = rowIndex+1, 0
		}
		rows[rowIndex], rowLength = append(rows[rowIndex], parameter.Key, parameter.Value), rowLength+1
		if rowLength != ParameterTableLineLength {
			rows[rowIndex] = append(rows[rowIndex], " ")
		}
	}
	t.AppendRows(rows)
	t.Render()
}

func PrintTable(title string, header table.Row, rows []table.Row) {
	t := buildDefaultTable()
	// Set the table's title
	if title != "" {
		t.SetTitle(title)
	}
	// Set the table's header and rows
	t.AppendHeader(header)
	t.AppendRows(rows)
	t.Render()
}
