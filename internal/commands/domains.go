package commands

import (
	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/output"
)

var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Manage container domains",
}

var domainsListCmd = &cobra.Command{
	Use:   "list <container>",
	Short: "List domains attached to a container",
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
		data, err := c.ListDomains(ctx(), id)
		if err != nil {
			return err
		}
		return output.PrintData("domains.list", data, jsonFlag, jqFlag)
	},
}

var domainsGetCmd = &cobra.Command{
	Use:   "get <container> <domain>",
	Short: "Get a single domain by name",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.GetDomain(ctx(), args[0], args[1])
		if err != nil {
			return err
		}
		return output.PrintData("domains.get", data, jsonFlag, jqFlag)
	},
}

var domainsDeleteCmd = &cobra.Command{
	Use:   "delete <container> <domain>",
	Short: "Delete a domain (DESTRUCTIVE — requires --yes)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		container, domain := args[0], args[1]
		if err := confirmDestructive("DELETE", "/api/v2/containers/"+container+"/domains/"+domain); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.DeleteDomain(ctx(), container, domain)
		if err != nil {
			return err
		}
		return output.PrintData("domains.delete", data, jsonFlag, jqFlag)
	},
}

func init() {
	domainsCmd.AddCommand(domainsListCmd, domainsGetCmd, domainsDeleteCmd)
}
