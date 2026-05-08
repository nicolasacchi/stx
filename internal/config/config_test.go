package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// withTempConfigDir points XDG_CONFIG_HOME at a fresh tmpdir for the duration of t.
func withTempConfigDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	return filepath.Join(tmp, "stx", "config.toml")
}

// clearStapeEnv removes any STAPE_* env vars that might leak from the host.
func clearStapeEnv(t *testing.T) {
	t.Helper()
	t.Setenv("STAPE_API_KEY", "")
	t.Setenv("STAPE_REGION", "")
	t.Setenv("STAPE_WORKSPACE", "")
}

func TestPath(t *testing.T) {
	withTempConfigDir(t)
	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if p == "" {
		t.Fatal("empty path")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	withTempConfigDir(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("missing file should return (nil, nil), got err: %v", err)
	}
	if cfg != nil {
		t.Errorf("expected nil cfg, got %+v", cfg)
	}
}

func TestSaveLoad_Roundtrip(t *testing.T) {
	withTempConfigDir(t)
	cfg := &Config{
		DefaultProject: "prod",
		Projects: map[string]Project{
			"prod": {APIKey: "k1", Region: "eu", Workspace: "ws-1", DefaultContainer: "abc"},
			"stg":  {APIKey: "k2", Region: "global"},
		},
	}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.DefaultProject != "prod" {
		t.Errorf("default project: got %q", loaded.DefaultProject)
	}
	if loaded.Projects["prod"].APIKey != "k1" {
		t.Errorf("prod api_key: got %q", loaded.Projects["prod"].APIKey)
	}
	if loaded.Projects["prod"].Workspace != "ws-1" {
		t.Errorf("prod workspace: got %q", loaded.Projects["prod"].Workspace)
	}
	if loaded.Projects["prod"].DefaultContainer != "abc" {
		t.Errorf("prod default_container: got %q", loaded.Projects["prod"].DefaultContainer)
	}
}

func TestLoadCredentials_FlagWins(t *testing.T) {
	withTempConfigDir(t)
	clearStapeEnv(t)
	t.Setenv("STAPE_API_KEY", "from-env")
	creds, err := LoadCredentials("from-flag", "global", "ws-flag", "")
	if err != nil {
		t.Fatal(err)
	}
	if creds.APIKey != "from-flag" {
		t.Errorf("flag should win; got %q", creds.APIKey)
	}
	if creds.Region != "global" {
		t.Errorf("region flag should win; got %q", creds.Region)
	}
	if creds.Workspace != "ws-flag" {
		t.Errorf("workspace flag should win; got %q", creds.Workspace)
	}
}

func TestLoadCredentials_EnvFallback(t *testing.T) {
	withTempConfigDir(t)
	clearStapeEnv(t)
	t.Setenv("STAPE_API_KEY", "from-env")
	t.Setenv("STAPE_REGION", "global")
	creds, err := LoadCredentials("", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if creds.APIKey != "from-env" {
		t.Errorf("env fallback failed; got %q", creds.APIKey)
	}
	if creds.Region != "global" {
		t.Errorf("region env failed; got %q", creds.Region)
	}
}

func TestLoadCredentials_PerFieldIndependent(t *testing.T) {
	// env api_key + config region must merge cleanly
	withTempConfigDir(t)
	clearStapeEnv(t)
	t.Setenv("STAPE_API_KEY", "from-env")
	cfg := &Config{
		DefaultProject: "prod",
		Projects:       map[string]Project{"prod": {APIKey: "from-config", Region: "global", Workspace: "from-config-ws"}},
	}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	creds, err := LoadCredentials("", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	// env wins for api_key; config used for region + workspace
	if creds.APIKey != "from-env" {
		t.Errorf("api_key: env should win; got %q", creds.APIKey)
	}
	if creds.Region != "global" {
		t.Errorf("region: config fallback; got %q", creds.Region)
	}
	if creds.Workspace != "from-config-ws" {
		t.Errorf("workspace: config fallback; got %q", creds.Workspace)
	}
}

func TestLoadCredentials_DefaultRegion(t *testing.T) {
	withTempConfigDir(t)
	clearStapeEnv(t)
	t.Setenv("STAPE_API_KEY", "k")
	creds, err := LoadCredentials("", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if creds.Region != "eu" {
		t.Errorf("default region should be eu; got %q", creds.Region)
	}
}

func TestLoadCredentials_InvalidRegion(t *testing.T) {
	withTempConfigDir(t)
	clearStapeEnv(t)
	t.Setenv("STAPE_API_KEY", "k")
	_, err := LoadCredentials("", "narnia", "", "")
	if err == nil {
		t.Fatal("expected error for invalid region")
	}
}

func TestLoadCredentials_NoCreds(t *testing.T) {
	withTempConfigDir(t)
	clearStapeEnv(t)
	_, err := LoadCredentials("", "", "", "")
	if !errors.Is(err, ErrNoCredentials) {
		t.Errorf("expected ErrNoCredentials, got %v", err)
	}
}

func TestLoadCredentials_NamedProject(t *testing.T) {
	withTempConfigDir(t)
	clearStapeEnv(t)
	cfg := &Config{
		Projects: map[string]Project{
			"prod": {APIKey: "k1", Region: "eu"},
			"stg":  {APIKey: "k2", Region: "global"},
		},
	}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	creds, err := LoadCredentials("", "", "", "stg")
	if err != nil {
		t.Fatal(err)
	}
	if creds.APIKey != "k2" {
		t.Errorf("named project: got %q", creds.APIKey)
	}
}

func TestLoadCredentials_UnknownNamedProject(t *testing.T) {
	withTempConfigDir(t)
	clearStapeEnv(t)
	cfg := &Config{Projects: map[string]Project{"prod": {APIKey: "k"}}}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	_, err := LoadCredentials("", "", "", "stg")
	if err == nil {
		t.Fatal("expected error for unknown project")
	}
}

func TestSave_Mode0600(t *testing.T) {
	path := withTempConfigDir(t)
	cfg := &Config{DefaultProject: "prod", Projects: map[string]Project{"prod": {APIKey: "k"}}}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("config file mode: got %v, want 0600", info.Mode().Perm())
	}
}
