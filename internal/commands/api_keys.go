package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/output"
)

var apiKeysCmd = &cobra.Command{
	Use:   "api-keys",
	Short: "Manage account API keys",
}

var apiKeysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your API keys (metadata only — secret values are NOT echoed)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.ListAPIKeys(ctx())
		if err != nil {
			return err
		}
		return output.PrintData("api-keys.list", data, jsonFlag, jqFlag)
	},
}

var apiKeysCreateName string

var apiKeysCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API key (DESTRUCTIVE — requires --yes; secret shown ONCE)",
	Long: `Create a new API key under your account.

The response includes the new key's secret value — Stape shows it ONLY ONCE.
Save it immediately (e.g. into 1Password). Subsequent 'api-keys list' calls
return only metadata (id, name), not the secret.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiKeysCreateName == "" {
			return fmt.Errorf("--name required")
		}
		if err := confirmDestructive("POST", "/api/v2/users/api-key"); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.CreateAPIKey(ctx(), apiKeysCreateName)
		if err != nil {
			return err
		}
		// Render normally; if on a TTY, prepend a save-now banner.
		if output.IsTTY() && jqFlag == "" {
			fmt.Fprintln(os.Stderr, "============================================================")
			fmt.Fprintln(os.Stderr, " save the key NOW — Stape will not show it again")
			fmt.Fprintln(os.Stderr, "============================================================")
		}
		return output.PrintData("api-keys.create", data, jsonFlag, jqFlag)
	},
}

var apiKeysDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Revoke an API key (DESTRUCTIVE — requires --yes)",
	Long: `Revoke an API key by its ID.

WARNING: do NOT delete the key currently configured in ~/.config/stx/config.toml
or you'll lock yourself out. The CLI cannot tell which key it's currently using
(the API never echoes the secret), so this protection is your responsibility.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if err := confirmDestructive("DELETE", "/api/v2/users/api-key/"+id); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.DeleteAPIKey(ctx(), id)
		if err != nil {
			return err
		}
		return output.PrintData("api-keys.delete", data, jsonFlag, jqFlag)
	},
}

func init() {
	apiKeysCreateCmd.Flags().StringVar(&apiKeysCreateName, "name", "", "Descriptive name for the new key (required)")
	apiKeysCmd.AddCommand(apiKeysListCmd, apiKeysCreateCmd, apiKeysDeleteCmd)
}
