//go:build ignore

// refresh.go — pull the OpenAPI spec from Stape's Swagger UI HTML and
// commit it to internal/openapi/spec.json.
//
// Stape does NOT expose a public /openapi.json endpoint; the spec is
// embedded inside the Swagger UI page as a <script id="swagger-data">
// JSON blob with shape {spec: {...}}.
//
// Usage:
//   go run ./internal/openapi/refresh.go
//
// Manual fallback if the scrape ever breaks:
//   1. Open https://api.app.eu.stape.io/api/doc in a browser
//   2. View source, find <script id="swagger-data">{...}</script>
//   3. Extract the .spec sub-object and write it (pretty-printed) to
//      internal/openapi/spec.json by hand
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	specURL = "https://api.app.eu.stape.io/api/doc"
	outPath = "internal/openapi/spec.json"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "refresh-spec failed:", err)
		os.Exit(1)
	}
}

func run() error {
	resp, err := http.Get(specURL)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", specURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("fetch %s: status %d (Stape may have changed Swagger UI)", specURL, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`<script[^>]+id="swagger-data"[^>]*>([\s\S]+?)</script>`)
	m := re.FindSubmatch(body)
	if m == nil {
		return fmt.Errorf(`no <script id="swagger-data"> tag at %s — Stape may have changed Swagger UI; paste spec manually into %s`, specURL, outPath)
	}

	var wrapper struct {
		Spec map[string]any `json:"spec"`
	}
	if err := json.Unmarshal(m[1], &wrapper); err != nil {
		return fmt.Errorf("parse swagger-data JSON: %w", err)
	}
	if wrapper.Spec == nil {
		return fmt.Errorf("swagger-data has no .spec key")
	}

	info, _ := wrapper.Spec["info"].(map[string]any)
	paths, _ := wrapper.Spec["paths"].(map[string]any)
	if info == nil || len(paths) == 0 {
		return fmt.Errorf("spec sanity-check failed (info=%v, paths=%d)", info != nil, len(paths))
	}
	if v, _ := info["version"].(string); !strings.HasPrefix(v, "2.") {
		fmt.Fprintf(os.Stderr, "warning: spec version %q is not 2.x — review changes carefully\n", v)
	}

	out, err := json.MarshalIndent(wrapper.Spec, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(outPath, out, 0644); err != nil {
		return err
	}

	fmt.Printf("wrote %s — %d bytes, %d endpoints, version=%v\n", outPath, len(out), len(paths), info["version"])
	return nil
}
