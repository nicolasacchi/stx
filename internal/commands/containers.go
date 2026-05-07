package commands

import (
	"github.com/spf13/cobra"

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

func init() {
	containersCmd.AddCommand(containersListCmd, containersGetCmd)
}
