package commands

import (
	"context"
	"fmt"
	"os"

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

	rootCmd.AddCommand(containersCmd, configCmd)
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
	c := client.New(creds.APIKey, creds.Region, creds.Workspace, verboseFlag)
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
