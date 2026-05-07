// Package output handles JSON / table rendering and gjson filtering.
//
// Convention (matches ddx/jx/gx):
//   - TTY default: table (when registered for the command)
//   - Pipe default: JSON
//   - --json: forces JSON on TTY
//   - --jq <gjson>: filters; applied only on the JSON path
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/tidwall/gjson"
)

// PrintData writes data to stdout as JSON, optionally filtered with gjson.
//
// command is a dotted lookup key reserved for future table rendering
// (e.g. "containers.list"); for v0 we always emit JSON.
func PrintData(command string, data []byte, jsonMode bool, jqFilter string) error {
	_ = command   // reserved
	_ = jsonMode  // reserved (will gate table render later)

	if jqFilter != "" {
		filtered, err := ApplyFilter(data, jqFilter)
		if err != nil {
			return err
		}
		data = filtered
	}
	return printJSON(os.Stdout, data)
}

// ApplyFilter runs a gjson expression against data.
// gjson syntax — NOT real jq. Examples:
//   "0.identifier"          first element's identifier
//   "#.identifier"          all identifiers (array)
//   "#.{id:identifier,name:name}"   project per element
func ApplyFilter(data []byte, expr string) ([]byte, error) {
	res := gjson.GetBytes(data, expr)
	if !res.Exists() {
		// Return JSON null for missing paths (matches ddx behavior)
		return []byte("null"), nil
	}
	return []byte(res.Raw), nil
}

func printJSON(w io.Writer, data []byte) error {
	// Re-indent for readability when output goes to a TTY
	if isTTY(w) {
		var pretty bytes
		if err := jsonIndent(&pretty, data); err == nil {
			_, err := fmt.Fprintln(w, pretty.String())
			return err
		}
	}
	_, err := w.Write(append(data, '\n'))
	return err
}

// IsTTY returns true when stdout is a terminal (pipes/redirects -> false).
func IsTTY() bool { return isTTY(os.Stdout) }

func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// --- thin wrappers to avoid an extra import in the test surface ---

type bytes struct{ buf []byte }

func (b *bytes) Write(p []byte) (int, error) { b.buf = append(b.buf, p...); return len(p), nil }
func (b *bytes) String() string               { return string(b.buf) }

func jsonIndent(dst *bytes, src []byte) error {
	var v any
	if err := json.Unmarshal(src, &v); err != nil {
		return err
	}
	enc := json.NewEncoder(dst)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
