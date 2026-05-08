package output

import (
	"strings"
	"testing"
)

func TestApplyFilter_SimpleField(t *testing.T) {
	data := []byte(`{"body":{"total":42,"items":[{"name":"a"},{"name":"b"}]}}`)
	out, err := ApplyFilter(data, "body.total")
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "42" {
		t.Errorf("got %q", out)
	}
}

func TestApplyFilter_ArrayCount(t *testing.T) {
	data := []byte(`{"body":{"items":[1,2,3]}}`)
	out, _ := ApplyFilter(data, "body.items.#")
	if string(out) != "3" {
		t.Errorf("got %q", out)
	}
}

func TestApplyFilter_ArrayProjection(t *testing.T) {
	data := []byte(`{"body":{"items":[{"name":"a"},{"name":"b"}]}}`)
	out, _ := ApplyFilter(data, "body.items.#.name")
	if string(out) != `["a","b"]` {
		t.Errorf("got %q", out)
	}
}

func TestApplyFilter_NestedObject(t *testing.T) {
	data := []byte(`{"body":{"items":[{"id":1,"name":"a"}]}}`)
	out, _ := ApplyFilter(data, "body.items.#.{i:id,n:name}")
	if !strings.Contains(string(out), `"i":1`) {
		t.Errorf("got %q", out)
	}
}

func TestApplyFilter_MissingPath_ReturnsNull(t *testing.T) {
	data := []byte(`{"x":1}`)
	out, err := ApplyFilter(data, "y.z")
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "null" {
		t.Errorf("got %q, want null", out)
	}
}

func TestApplyFilter_FirstElement(t *testing.T) {
	data := []byte(`{"body":[{"identifier":"abc"},{"identifier":"def"}]}`)
	out, _ := ApplyFilter(data, "body.0.identifier")
	if string(out) != `"abc"` {
		t.Errorf("got %q", out)
	}
}

func TestPrintData_JSONMode_NoFilter(t *testing.T) {
	// PrintData prints to os.Stdout — we can't capture cleanly without flag wrangling.
	// At minimum, exercise the code path without panic.
	if err := PrintData("test", []byte(`{"x":1}`), true, ""); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestPrintData_WithFilter(t *testing.T) {
	if err := PrintData("test", []byte(`{"body":{"x":1}}`), true, "body.x"); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestPrintData_UnknownCommand_FallsBackToJSON(t *testing.T) {
	// Command "no-such-thing" has no column registration, should fall through to JSON.
	if err := PrintData("no-such-thing", []byte(`{"x":1}`), false, ""); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestIsTTY_NotForPipe(t *testing.T) {
	// IsTTY checks os.Stdout. In a test runner this is generally not a TTY.
	// We don't assert a specific value (depends on environment), but exercise the code.
	_ = IsTTY()
}
