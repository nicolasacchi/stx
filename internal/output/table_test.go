package output

import (
	"errors"
	"testing"
)

func TestPrintTable_NoLayout(t *testing.T) {
	err := printTable("nonexistent.command", []byte(`{}`))
	if !errors.Is(err, errNoTable) {
		t.Errorf("expected errNoTable, got %v", err)
	}
}

func TestPrintTable_MissingRowsPath(t *testing.T) {
	// monitoring.logs.detailed has a column map; if body.items is missing, expect explicit error
	err := printTable("monitoring.logs.detailed", []byte(`{"unrelated": 1}`))
	if err == nil || errors.Is(err, errNoTable) {
		t.Errorf("expected real error for missing rows, got %v", err)
	}
}

func TestPrintTable_RendersWithEmptyArray(t *testing.T) {
	// monitoring.logs.detailed with body.items=[] should print "(no rows)" without error
	err := printTable("monitoring.logs.detailed", []byte(`{"body":{"items":[]}}`))
	if err != nil {
		t.Errorf("expected no error for empty rows, got %v", err)
	}
}

func TestPrintTable_RendersWithRows(t *testing.T) {
	// Smoke test: actual rendering on stdout. Don't assert exact output.
	data := []byte(`{"body":{"items":[
		{"Date":1700000000,"Status":"204","Method":"POST","Platform":"GA4","Event":"PageView","URL":"https://x.com/g/collect"}
	]}}`)
	if err := printTable("monitoring.logs.detailed", data); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
