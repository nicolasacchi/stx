package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/tidwall/gjson"
)

// ColumnDef defines one table column for one command.
// Key is a flat key into the row map (no dotted paths — pre-flatten in the
// client method or use a different command-level path key like "body.items").
type ColumnDef struct {
	Header string
	Key    string
	Format FormatFunc
}

// commandColumns maps "<group>.<leaf>" command keys to column layouts.
//
// To register a new command for table rendering: add an entry below.
// Commands without an entry fall through to the JSON output path.
//
// Pre-flatten contract: the renderer reads `body.items[]` (Stape's response
// envelope) by default. If a command's payload is at the root (analytics
// endpoints return arrays at the root), set commandRowsPath[command] = "@this".
var commandColumns = map[string][]ColumnDef{
	// Field names below match the LIVE Stape API response (PascalCase, hyphenated),
	// which is DIFFERENT from the CSV export field names (snake_case).
	// Verified against /api/v2/containers/{id}/monitoring/logs/outgoing/detailed.
	"monitoring.logs.detailed": {
		{Header: "TIME", Key: "Date", Format: FormatEpochSeconds},
		{Header: "STATUS", Key: "Status"},
		{Header: "METHOD", Key: "Method"},
		{Header: "PLATFORM", Key: "Platform"},
		{Header: "EVENT", Key: "Event"},
		{Header: "HOST", Key: "URL", Format: FormatURLHost},
	},
}

// commandRowsPath overrides the gjson path used to extract the rows array.
// Default is "body.items" (Stape's envelope).
var commandRowsPath = map[string]string{}

// errNoTable signals the caller (output.PrintData) to fall back to JSON.
var errNoTable = errors.New("no table layout")

func printTable(command string, data []byte) error {
	cols, ok := commandColumns[command]
	if !ok {
		return errNoTable
	}

	rowsPath := commandRowsPath[command]
	if rowsPath == "" {
		rowsPath = "body.items"
	}

	rows := gjson.GetBytes(data, rowsPath)
	if !rows.Exists() {
		return fmt.Errorf("no rows at %q", rowsPath)
	}

	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.SetStyle(table.StyleLight)
	tw.Style().Format.Header = text.FormatUpper

	headers := make(table.Row, len(cols))
	for i, c := range cols {
		headers[i] = c.Header
	}
	tw.AppendHeader(headers)

	rendered := 0
	rows.ForEach(func(_, row gjson.Result) bool {
		out := make(table.Row, len(cols))
		for i, c := range cols {
			cell := row.Get(c.Key)
			var v any
			if !cell.Exists() {
				v = ""
			} else {
				v = unwrap(cell)
			}
			if c.Format != nil {
				out[i] = c.Format(v)
			} else {
				out[i] = v
			}
		}
		tw.AppendRow(out)
		rendered++
		return true
	})

	if rendered == 0 {
		fmt.Println("(no rows)")
		return nil
	}

	tw.Render()
	return nil
}

func unwrap(r gjson.Result) any {
	switch r.Type {
	case gjson.Number:
		return r.Num
	case gjson.String:
		return r.Str
	case gjson.True:
		return true
	case gjson.False:
		return false
	case gjson.Null:
		return nil
	default:
		var v any
		_ = json.Unmarshal([]byte(r.Raw), &v)
		return v
	}
}
