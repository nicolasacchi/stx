package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/client"
	"github.com/nicolasacchi/stx/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage ~/.config/stx/config.toml",
}

var (
	configAddName             string
	configAddAPIKey           string
	configAddRegion           string
	configAddWorkspace        string
	configAddDefaultContainer string
)

var configAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add or replace a project block",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if cfg == nil {
			cfg = &config.Config{Projects: map[string]config.Project{}}
		}
		region := configAddRegion
		if region == "" {
			region = "eu"
		}
		cfg.Projects[name] = config.Project{
			APIKey:           configAddAPIKey,
			Region:           region,
			Workspace:        configAddWorkspace,
			DefaultContainer: configAddDefaultContainer,
		}
		if cfg.DefaultProject == "" {
			cfg.DefaultProject = name
		}
		if err := config.Save(cfg); err != nil {
			return err
		}
		path, _ := config.Path()
		fmt.Printf("saved project %q to %s\n", name, path)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured projects",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if cfg == nil || len(cfg.Projects) == 0 {
			fmt.Println("(no projects configured)")
			return nil
		}
		names := make([]string, 0, len(cfg.Projects))
		for n := range cfg.Projects {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			marker := " "
			if n == cfg.DefaultProject {
				marker = "*"
			}
			p := cfg.Projects[n]
			fmt.Printf("%s %-20s region=%-6s workspace=%s\n", marker, n, p.Region, redact(p.Workspace))
		}
		return nil
	},
}

var configUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set default project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if cfg == nil {
			return fmt.Errorf("no config file; run 'stx config add' first")
		}
		if _, ok := cfg.Projects[name]; !ok {
			return fmt.Errorf("project %q not found", name)
		}
		cfg.DefaultProject = name
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("default project: %s\n", name)
		return nil
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a project block",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if cfg == nil {
			return fmt.Errorf("no config file")
		}
		if _, ok := cfg.Projects[name]; !ok {
			return fmt.Errorf("project %q not found", name)
		}
		delete(cfg.Projects, name)
		if cfg.DefaultProject == name {
			cfg.DefaultProject = ""
			for n := range cfg.Projects {
				cfg.DefaultProject = n
				break
			}
		}
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("removed project %q\n", name)
		return nil
	},
}

var configCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show resolved credentials (API key redacted)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := config.LoadCredentials(apiKeyFlag, regionFlag, workspaceFlag, projectFlag)
		if err != nil {
			return err
		}
		fmt.Printf("api_key:           %s\n", redact(creds.APIKey))
		fmt.Printf("region:            %s\n", creds.Region)
		fmt.Printf("workspace:         %s\n", emptyOrValue(creds.Workspace, "(none)"))
		fmt.Printf("default_container: %s\n", emptyOrValue(creds.DefaultContainer, "(none)"))
		return nil
	},
}

var configDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Verify credentials by calling GET /api/v2/users",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, creds, err := getClient()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "→ %s (region=%s, workspace=%s)\n", c.BaseURL(), creds.Region, emptyOrValue(creds.Workspace, "none"))
		data, err := c.ListUsers(ctx())
		if err != nil {
			fmt.Println("status: FAIL")
			return err
		}
		var probe struct {
			Items []struct {
				Email    string `json:"email"`
				Username string `json:"username"`
			} `json:"items"`
		}
		_ = json.Unmarshal(data, &probe)
		fmt.Println("status: ok")
		fmt.Printf("response: %d bytes\n", len(data))
		if len(probe.Items) > 0 {
			fmt.Printf("first user: %s / %s\n", probe.Items[0].Email, probe.Items[0].Username)
		}
		return nil
	},
}

func init() {
	configAddCmd.Flags().StringVar(&configAddAPIKey, "api-key", "", "Stape API key (required)")
	configAddCmd.Flags().StringVar(&configAddRegion, "region", "eu", "Stape region: eu or global")
	configAddCmd.Flags().StringVar(&configAddWorkspace, "workspace", "", "Workspace UUID (optional)")
	configAddCmd.Flags().StringVar(&configAddDefaultContainer, "default-container", "", "Default container identifier (optional)")
	_ = configAddCmd.MarkFlagRequired("api-key")

	configCmd.AddCommand(configAddCmd, configListCmd, configUseCmd, configRemoveCmd, configCurrentCmd, configDoctorCmd)
	_ = client.MaxRetries // keep import live
}

func redact(s string) string {
	if s == "" {
		return "(none)"
	}
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "..." + s[len(s)-4:]
}

func emptyOrValue(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
