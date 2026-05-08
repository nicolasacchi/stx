package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/output"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Agency sub-user management",
	Long: `Sub-user management endpoints. Note: most return 403 on personal
(non-agency) accounts. The 'export-csv' command returns text/csv (not JSON)
and bypasses --jq.`,
}

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List users",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.ListUsers(ctx())
		if err != nil {
			return err
		}
		return output.PrintData("users.list", data, jsonFlag, jqFlag)
	},
}

var usersGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a single user by identifier",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.GetUser(ctx(), args[0])
		if err != nil {
			return err
		}
		return output.PrintData("users.get", data, jsonFlag, jqFlag)
	},
}

// --- create ---

var (
	usersCreateUsername      string
	usersCreatePassword      string
	usersCreatePasswordStdin bool
	usersCreateNameFirst     string
	usersCreateNameLast      string
	usersCreateSendEmail     bool
)

var usersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a sub-user (DESTRUCTIVE — requires --yes)",
	Long: `Create a sub-user under your agency account.

Required:
  --username <name>
  one of: --password <plaintext> | --password-stdin

--password is visible in the process list (ps); prefer --password-stdin
for scripts. Reads stdin until EOF and trims trailing newline.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if usersCreateUsername == "" {
			return fmt.Errorf("--username required")
		}
		if usersCreatePassword == "" && !usersCreatePasswordStdin {
			return fmt.Errorf("provide --password <pass> or --password-stdin")
		}
		if usersCreatePassword != "" && usersCreatePasswordStdin {
			return fmt.Errorf("--password and --password-stdin are mutually exclusive")
		}
		if usersCreatePasswordStdin {
			raw, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read --password-stdin: %w", err)
			}
			usersCreatePassword = strings.TrimRight(string(raw), "\r\n")
			if usersCreatePassword == "" {
				return fmt.Errorf("empty password from stdin")
			}
		}
		if err := confirmDestructive("POST", "/api/v2/users"); err != nil {
			return err
		}
		body := map[string]any{
			"username": usersCreateUsername,
			"password": usersCreatePassword,
		}
		if usersCreateNameFirst != "" {
			body["nameFirst"] = usersCreateNameFirst
		}
		if usersCreateNameLast != "" {
			body["nameLast"] = usersCreateNameLast
		}
		if usersCreateSendEmail {
			body["sendEmail"] = true
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.CreateUser(ctx(), body)
		if err != nil {
			return err
		}
		return output.PrintData("users.create", data, jsonFlag, jqFlag)
	},
}

// --- attach ---

var (
	usersAttachEmail         string
	usersAttachHasNoProducts bool
	usersAttachProductName   string
)

var usersAttachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach an existing user to your agency (DESTRUCTIVE — requires --yes)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if usersAttachEmail == "" {
			return fmt.Errorf("--email required")
		}
		if err := confirmDestructive("POST", "/api/v2/users/attach-user"); err != nil {
			return err
		}
		body := map[string]any{
			"email":          usersAttachEmail,
			"hasNoProducts":  usersAttachHasNoProducts,
		}
		if usersAttachProductName != "" {
			body["productName"] = usersAttachProductName
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.AttachUser(ctx(), body)
		if err != nil {
			return err
		}
		return output.PrintData("users.attach", data, jsonFlag, jqFlag)
	},
}

// --- detach ---

var usersDetachCmd = &cobra.Command{
	Use:   "detach <id>",
	Short: "Detach a sub-user (DESTRUCTIVE — requires --yes)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if err := confirmDestructive("POST", "/api/v2/users/"+id+"/detach-user"); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.DetachUser(ctx(), id)
		if err != nil {
			return err
		}
		return output.PrintData("users.detach", data, jsonFlag, jqFlag)
	},
}

// --- export-csv (CSV bypass) ---

var usersExportOut string

var usersExportCSVCmd = &cobra.Command{
	Use:   "export-csv",
	Short: "Export users as CSV (NOT JSON — bypasses --jq)",
	Long: `Returns Content-Type: text/csv. The output bypasses the JSON
formatting layer entirely.

  --out <file>    write to file (mode 0644)
  (default)       write raw bytes to stdout

--jq is not supported on this command and will return an error.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if jqFlag != "" {
			return fmt.Errorf("--jq not supported on users export-csv (response is text/csv)")
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.ExportUsersCSV(ctx())
		if err != nil {
			return err
		}
		if usersExportOut != "" {
			return os.WriteFile(usersExportOut, data, 0644)
		}
		_, err = os.Stdout.Write(data)
		return err
	},
}

func init() {
	usersCreateCmd.Flags().StringVar(&usersCreateUsername, "username", "", "Username (required)")
	usersCreateCmd.Flags().StringVar(&usersCreatePassword, "password", "", "Password (visible in ps; prefer --password-stdin)")
	usersCreateCmd.Flags().BoolVar(&usersCreatePasswordStdin, "password-stdin", false, "Read password from stdin")
	usersCreateCmd.Flags().StringVar(&usersCreateNameFirst, "name-first", "", "First name (optional)")
	usersCreateCmd.Flags().StringVar(&usersCreateNameLast, "name-last", "", "Last name (optional)")
	usersCreateCmd.Flags().BoolVar(&usersCreateSendEmail, "send-email", false, "Send welcome email to the new user")

	usersAttachCmd.Flags().StringVar(&usersAttachEmail, "email", "", "Email of the user to attach (required)")
	usersAttachCmd.Flags().BoolVar(&usersAttachHasNoProducts, "has-no-products", false, "Set hasNoProducts=true (required by API)")
	usersAttachCmd.Flags().StringVar(&usersAttachProductName, "product-name", "", "Product name (optional)")

	usersExportCSVCmd.Flags().StringVar(&usersExportOut, "out", "", "Write CSV to file instead of stdout")

	usersCmd.AddCommand(
		usersListCmd, usersGetCmd,
		usersCreateCmd, usersAttachCmd, usersDetachCmd,
		usersExportCSVCmd,
	)

	_ = json.Marshal // keep import live for future
}
