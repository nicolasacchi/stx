package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/client"
	"github.com/nicolasacchi/stx/internal/config"
)

// Persistent flag values. cobra binds these in init().
var (
	apiKeyFlag    string
	regionFlag    string
	workspaceFlag string
	projectFlag   string
	jsonFlag      bool
	jqFlag        string
	verboseFlag   bool
	limitFlag     int
	timeoutFlag   string
	yesFlag       bool
)

var rootCmd = &cobra.Command{
	Use:           "stx",
	Short:         "stx — Stape Explorer CLI",
	Long:          "Read-and-write CLI for the Stape API. JSON output, gjson --jq filters, multi-region, multi-project, multi-workspace.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// SetVersion is called by main.go before Execute.
func SetVersion(v string) {
	rootCmd.Version = v
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "Stape API key (overrides STAPE_API_KEY env / config)")
	rootCmd.PersistentFlags().StringVar(&regionFlag, "region", "", "Stape region: eu (default) or global")
	rootCmd.PersistentFlags().StringVar(&workspaceFlag, "workspace", "", "Stape workspace UUID (X-WORKSPACE header)")
	rootCmd.PersistentFlags().StringVar(&projectFlag, "project", "", "Named project block from ~/.config/stx/config.toml")
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Force JSON output even on TTY")
	rootCmd.PersistentFlags().StringVar(&jqFlag, "jq", "", "gjson filter expression (NOT real jq)")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Log HTTP requests to stderr")
	rootCmd.PersistentFlags().IntVar(&limitFlag, "limit", 0, "Cap result count (0 = API default)")
	rootCmd.PersistentFlags().StringVar(&timeoutFlag, "timeout", "30s", "HTTP timeout (e.g. 60s, 2m)")
	rootCmd.PersistentFlags().BoolVar(&yesFlag, "yes", false, "Skip confirmation prompt for destructive commands")

	rootCmd.AddCommand(
		containersCmd, configCmd, analyticsCmd, monitoringCmd, domainsCmd, powerUpsCmd,
		customLoaderCmd, proxyFilesCmd, schedulesCmd,
	)
}

// confirmDestructive prints a confirmation prompt and returns nil iff --yes was set.
// Used by delete/transfer/update commands.
func confirmDestructive(method, path string) error {
	if yesFlag {
		return nil
	}
	fmt.Fprintf(os.Stderr, "about to %s %s\nre-run with --yes to confirm\n", method, path)
	return fmt.Errorf("confirmation required")
}

// readBodyFile reads a JSON file from disk and returns the raw bytes.
// Validates by attempting to parse as JSON; rejects empty files.
func readBodyFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("--from-file path required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("%s: empty file", path)
	}
	var probe any
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, fmt.Errorf("%s: invalid JSON: %w", path, err)
	}
	return data, nil
}

// getClient resolves credentials and constructs the HTTP client.
func getClient() (*client.Client, *config.Credentials, error) {
	creds, err := config.LoadCredentials(apiKeyFlag, regionFlag, workspaceFlag, projectFlag)
	if err != nil {
		return nil, nil, err
	}
	if verboseFlag {
		client.SetVerboseDest(os.Stderr)
	}
	timeout, err := time.ParseDuration(timeoutFlag)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid --timeout %q: %w", timeoutFlag, err)
	}
	c := client.New(creds.APIKey, creds.Region, creds.Workspace, verboseFlag, timeout)
	return c, creds, nil
}

// ctx returns a request context. Reserved for future timeout flag.
func ctx() context.Context { return context.Background() }

// resolveContainer returns the explicit positional arg or the per-project default_container.
func resolveContainer(args []string, creds *config.Credentials) (string, error) {
	if len(args) > 0 && args[0] != "" {
		return args[0], nil
	}
	if creds != nil && creds.DefaultContainer != "" {
		return creds.DefaultContainer, nil
	}
	return "", fmt.Errorf("container ID required (positional arg or default_container in config)")
}
