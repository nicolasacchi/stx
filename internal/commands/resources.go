package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/client"
	"github.com/nicolasacchi/stx/internal/output"
)

var resourcesCmd = &cobra.Command{
	Use:   "resources <kind>",
	Short: "Look up enum / reference data by kind",
	Long: `Stape exposes 15 reference endpoints under /api/v2/resources/<kind>.
Use 'stx resources kinds' to print the full list.

Examples:
  stx resources container-zones                # available zones
  stx resources container-statuses             # status enum values
  stx resources container-domain-cdn-types     # cdn type enum`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kind := args[0]
		if kind == "kinds" {
			fmt.Println("Valid resource kinds:")
			for _, k := range client.ResourceKinds {
				fmt.Println("  " + k)
			}
			return nil
		}
		if !client.IsValidResourceKind(kind) {
			return fmt.Errorf("invalid resource kind %q\n\nvalid kinds:\n  %s", kind, strings.Join(client.ResourceKinds, "\n  "))
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.GetResource(ctx(), kind)
		if err != nil {
			return err
		}
		return output.PrintData("resources."+kind, data, jsonFlag, jqFlag)
	},
}
