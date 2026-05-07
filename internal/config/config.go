// Package config loads multi-project credentials from ~/.config/stx/config.toml
// and resolves them per-field (flag > env > project block > default_project).
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	EnvAPIKey    = "STAPE_API_KEY"
	EnvRegion    = "STAPE_REGION"
	EnvWorkspace = "STAPE_WORKSPACE"
)

// Credentials is the resolved set used by the HTTP client.
type Credentials struct {
	APIKey           string
	Region           string // "eu" or "global"
	Workspace        string // optional UUID, sent as X-WORKSPACE
	DefaultContainer string // optional, fallback when positional missing
}

// Project is one entry in config.toml under [projects.<name>].
type Project struct {
	APIKey           string `toml:"api_key"`
	Region           string `toml:"region"`
	Workspace        string `toml:"workspace"`
	DefaultContainer string `toml:"default_container"`
}

// Config is the on-disk shape of ~/.config/stx/config.toml.
type Config struct {
	DefaultProject string             `toml:"default_project"`
	Projects       map[string]Project `toml:"projects"`
}

// ErrNoCredentials is returned when neither flag, env, nor config provides an API key.
var ErrNoCredentials = errors.New("API key required: use --api-key flag, STAPE_API_KEY env var, or run 'stx config add <name>'")

// LoadCredentials resolves credentials per-field from flag > env > project > default_project.
// Each field resolves independently — STAPE_API_KEY env can pair with config.region.
func LoadCredentials(apiKeyFlag, regionFlag, workspaceFlag, projectFlag string) (*Credentials, error) {
	creds := &Credentials{}

	cfg, _ := Load() // missing config is OK; we'll fall through to flag/env

	// Pick a project block to read defaults from
	var proj *Project
	if cfg != nil {
		name := projectFlag
		if name == "" {
			name = cfg.DefaultProject
		}
		if name != "" {
			if p, ok := cfg.Projects[name]; ok {
				proj = &p
			} else if projectFlag != "" {
				return nil, fmt.Errorf("project %q not found in config; available: %v", projectFlag, projectNames(cfg.Projects))
			}
		}
	}

	// Resolve each field independently
	creds.APIKey = firstNonEmpty(apiKeyFlag, os.Getenv(EnvAPIKey), projectField(proj, func(p *Project) string { return p.APIKey }))
	creds.Region = firstNonEmpty(regionFlag, os.Getenv(EnvRegion), projectField(proj, func(p *Project) string { return p.Region }))
	creds.Workspace = firstNonEmpty(workspaceFlag, os.Getenv(EnvWorkspace), projectField(proj, func(p *Project) string { return p.Workspace }))
	if proj != nil {
		creds.DefaultContainer = proj.DefaultContainer
	}

	if creds.Region == "" {
		creds.Region = "eu"
	}
	if creds.Region != "eu" && creds.Region != "global" {
		return nil, fmt.Errorf("invalid region %q (must be eu or global)", creds.Region)
	}

	if creds.APIKey == "" {
		return nil, ErrNoCredentials
	}

	return creds, nil
}

// Load reads ~/.config/stx/config.toml. Missing file returns (nil, nil).
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if cfg.Projects == nil {
		cfg.Projects = map[string]Project{}
	}
	return &cfg, nil
}

// Save writes the config to ~/.config/stx/config.toml (creating dir if needed, mode 0600).
func Save(cfg *Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := toml.NewEncoder(f)
	return enc.Encode(cfg)
}

// Path returns the config file path: $XDG_CONFIG_HOME/stx/config.toml or ~/.config/stx/config.toml.
func Path() (string, error) {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return filepath.Join(v, "stx", "config.toml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "stx", "config.toml"), nil
}

func firstNonEmpty(s ...string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}

func projectField(p *Project, get func(*Project) string) string {
	if p == nil {
		return ""
	}
	return get(p)
}

func projectNames(m map[string]Project) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
