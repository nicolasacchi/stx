package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/client"
	"github.com/nicolasacchi/stx/internal/output"
)

var containersCmd = &cobra.Command{
	Use:   "containers",
	Short: "Manage Stape containers",
}

var containersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all containers in the workspace",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.ListContainers(ctx(), limitFlag)
		if err != nil {
			return err
		}
		return output.PrintData("containers.list", data, jsonFlag, jqFlag)
	},
}

var containersGetCmd = &cobra.Command{
	Use:   "get <container>",
	Short: "Get a single container by identifier",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, creds, err := getClient()
		if err != nil {
			return err
		}
		id, err := resolveContainer(args, creds)
		if err != nil {
			return err
		}
		data, err := c.GetContainer(ctx(), id)
		if err != nil {
			return err
		}
		return output.PrintData("containers.get", data, jsonFlag, jqFlag)
	},
}

// --- Create ---

var (
	containersCreateFromFile string
)

var containersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new container (DESTRUCTIVE — requires --yes)",
	Long: `Create a Stape container.

Body must satisfy ContainerCreateForm: required fields are
name, code (≥64 chars), codeSettings, anonymizeOptions, zone.
Pass the full payload via --from-file <path> (JSON).`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := confirmDestructive("POST", "/api/v2/containers"); err != nil {
			return err
		}
		body, err := readBodyFile(containersCreateFromFile)
		if err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		var raw json.RawMessage = body
		data, err := c.CreateContainer(ctx(), raw)
		if err != nil {
			return err
		}
		return output.PrintData("containers.create", data, jsonFlag, jqFlag)
	},
}

// --- Update ---

var (
	containersUpdateFromFile string
	containersUpdateName     string
)

var containersUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a container (DESTRUCTIVE — requires --yes; PUT is full-document replace)",
	Long: `Update a Stape container.

Two mutually-exclusive modes:
  --from-file <path>   PUT the JSON file contents directly.
  --name <new>          GET the container, replace name, PUT it back.

Stape PUT is full-document replace — partial PATCH is not supported.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if (containersUpdateFromFile == "") == (containersUpdateName == "") {
			return fmt.Errorf("specify exactly one of --from-file or --name")
		}
		if err := confirmDestructive("PUT", "/api/v2/containers/"+id); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		var body json.RawMessage
		if containersUpdateFromFile != "" {
			b, err := readBodyFile(containersUpdateFromFile)
			if err != nil {
				return err
			}
			body = b
		} else {
			// GET → mutate → PUT
			current, err := c.GetContainer(ctx(), id)
			if err != nil {
				return fmt.Errorf("fetch current: %w", err)
			}
			b, err := projectUpdateForm(current, containersUpdateName)
			if err != nil {
				return err
			}
			body = b
		}
		data, err := c.UpdateContainer(ctx(), id, body)
		if err != nil {
			return err
		}
		return output.PrintData("containers.update", data, jsonFlag, jqFlag)
	},
}

// projectUpdateForm extracts ContainerUpdateForm fields from a GetContainer
// response (which is wrapped in {"body": {...}}) and applies a name mutation.
func projectUpdateForm(getResp []byte, newName string) ([]byte, error) {
	var wrapper struct {
		Body map[string]any `json:"body"`
	}
	if err := json.Unmarshal(getResp, &wrapper); err != nil {
		return nil, fmt.Errorf("parse GET response: %w", err)
	}
	if wrapper.Body == nil {
		return nil, fmt.Errorf("GET response missing body")
	}
	form := map[string]any{}
	for _, k := range []string{
		"name", "code", "codeSettings", "anonymizeOptions",
		"geoHeaders", "xmlToJsonEnabled", "delayEnabled",
		"stapeUserIdHeaderEnabled", "botIndexEnabled", "userAgentHeaders",
		"serviceAccountCredentials",
	} {
		if v, ok := wrapper.Body[k]; ok && v != nil {
			form[k] = v
		}
	}
	if newName != "" {
		form["name"] = newName
	}
	return json.Marshal(form)
}

// --- Delete ---

var (
	containersDeleteReasonSetup  string
	containersDeleteReasonCancel string
)

var containersDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a container (DESTRUCTIVE — requires --yes)",
	Long: `Delete a Stape container.

By default sends an empty body (cancelReason is form-level optional in the spec).
If Stape rejects with 400, populate the reason arrays:

  stx containers delete <id> --yes \
    --reason-setup ga,other \
    --reason-cancel use,price`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if err := confirmDestructive("DELETE", "/api/v2/containers/"+id); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		var body any
		if containersDeleteReasonSetup != "" || containersDeleteReasonCancel != "" {
			body = client.SubscriptionCancelForm{
				CancelReason: &client.SubscriptionCancelReason{
					Setup:  parseReasons(containersDeleteReasonSetup),
					Cancel: parseReasons(containersDeleteReasonCancel),
				},
			}
		}
		data, err := c.DeleteContainer(ctx(), id, body)
		if err != nil {
			// Helpful hint for the canonical 400
			if strings.Contains(err.Error(), "400") && body == nil {
				return fmt.Errorf("%w\nhint: Stape may require cancel reasons. Try:\n  stx containers delete %s --yes --reason-setup ga,other --reason-cancel use,price", err, id)
			}
			return err
		}
		return output.PrintData("containers.delete", data, jsonFlag, jqFlag)
	},
}

func parseReasons(csv string) []client.OptionForm {
	if csv == "" {
		return []client.OptionForm{}
	}
	parts := strings.Split(csv, ",")
	out := make([]client.OptionForm, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, client.OptionForm{Type: p})
	}
	return out
}

// --- Transfer ---

var containersTransferEmail string

var containersTransferCmd = &cobra.Command{
	Use:   "transfer <id>",
	Short: "Transfer a container to another user (DESTRUCTIVE — requires --yes)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if containersTransferEmail == "" {
			return fmt.Errorf("--email required")
		}
		if err := confirmDestructive("PUT", "/api/v2/containers/"+id+"/transfer"); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.TransferContainer(ctx(), id, containersTransferEmail)
		if err != nil {
			return err
		}
		return output.PrintData("containers.transfer", data, jsonFlag, jqFlag)
	},
}

func init() {
	containersCreateCmd.Flags().StringVar(&containersCreateFromFile, "from-file", "", "Path to JSON body (ContainerCreateForm)")
	_ = containersCreateCmd.MarkFlagRequired("from-file")

	containersUpdateCmd.Flags().StringVar(&containersUpdateFromFile, "from-file", "", "Path to JSON body (ContainerUpdateForm) — full-document replace")
	containersUpdateCmd.Flags().StringVar(&containersUpdateName, "name", "", "New name (GET-then-PUT mutation)")

	containersDeleteCmd.Flags().StringVar(&containersDeleteReasonSetup, "reason-setup", "", "Comma-separated cancelReason.setup types (e.g. ga,other)")
	containersDeleteCmd.Flags().StringVar(&containersDeleteReasonCancel, "reason-cancel", "", "Comma-separated cancelReason.cancel types (e.g. use,price)")

	containersTransferCmd.Flags().StringVar(&containersTransferEmail, "email", "", "Email of the new owner (required)")
	_ = containersTransferCmd.MarkFlagRequired("email")

	containersCmd.AddCommand(
		containersListCmd, containersGetCmd,
		containersCreateCmd, containersUpdateCmd, containersDeleteCmd, containersTransferCmd,
	)
}
