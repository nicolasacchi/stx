package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/output"
)

// Shared domain-form flags (used by create / update / validate).
var (
	domainName           string
	domainCdnType        string
	domainConnectionType string
	domainUseCnameRecord bool
	domainUseARecord     bool
)

func domainFormBody() (map[string]any, error) {
	if domainName == "" {
		return nil, fmt.Errorf("--name required")
	}
	if domainCdnType == "" {
		return nil, fmt.Errorf("--cdn-type required (stape|custom|none)")
	}
	switch domainCdnType {
	case "stape", "custom", "none":
	default:
		return nil, fmt.Errorf("--cdn-type must be one of: stape, custom, none")
	}
	body := map[string]any{
		"name":    domainName,
		"cdnType": domainCdnType,
	}
	if domainConnectionType != "" {
		switch domainConnectionType {
		case "stape", "entri":
		default:
			return nil, fmt.Errorf("--connection-type must be one of: stape, entri")
		}
		body["connectionType"] = domainConnectionType
	}
	if domainUseCnameRecord {
		body["useCnameRecord"] = true
	}
	if domainUseARecord {
		body["useARecord"] = true
	}
	return body, nil
}

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

var domainsCreateCmd = &cobra.Command{
	Use:   "create <container>",
	Short: "Add a domain to a container (DESTRUCTIVE — requires --yes)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		container := args[0]
		body, err := domainFormBody()
		if err != nil {
			return err
		}
		if err := confirmDestructive("POST", "/api/v2/containers/"+container+"/domains"); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.CreateDomain(ctx(), container, body)
		if err != nil {
			return err
		}
		return output.PrintData("domains.create", data, jsonFlag, jqFlag)
	},
}

var domainsUpdateCmd = &cobra.Command{
	Use:   "update <container> <domain-uuid>",
	Short: "Replace domain config (full PUT) — requires --yes",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		container, domain := args[0], args[1]
		body, err := domainFormBody()
		if err != nil {
			return err
		}
		if err := confirmDestructive("PUT", "/api/v2/containers/"+container+"/domains/"+domain); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.UpdateDomain(ctx(), container, domain, body)
		if err != nil {
			return err
		}
		return output.PrintData("domains.update", data, jsonFlag, jqFlag)
	},
}

var domainsValidateCmd = &cobra.Command{
	Use:   "validate <container>",
	Short: "Validate a domain config without persisting it",
	Long:  "Reports whether the proposed domain config would succeed. No state mutation; safe to run.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		container := args[0]
		body, err := domainFormBody()
		if err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.ValidateDomain(ctx(), container, body)
		if err != nil {
			return err
		}
		return output.PrintData("domains.validate", data, jsonFlag, jqFlag)
	},
}

var domainsRevalidateCmd = &cobra.Command{
	Use:   "revalidate <container> <domain-uuid>",
	Short: "Re-check DNS / cert / readiness for an existing domain",
	Long:  "Idempotent — does not modify the domain config. Safe to run.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		container, domain := args[0], args[1]
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.RevalidateDomain(ctx(), container, domain)
		if err != nil {
			return err
		}
		return output.PrintData("domains.revalidate", data, jsonFlag, jqFlag)
	},
}

func init() {
	for _, c := range []*cobra.Command{domainsCreateCmd, domainsUpdateCmd, domainsValidateCmd} {
		c.Flags().StringVar(&domainName, "name", "", "Domain hostname (required)")
		c.Flags().StringVar(&domainCdnType, "cdn-type", "", "stape|custom|none (required)")
		c.Flags().StringVar(&domainConnectionType, "connection-type", "", "stape|entri (optional)")
		c.Flags().BoolVar(&domainUseCnameRecord, "use-cname-record", false, "Use CNAME DNS record")
		c.Flags().BoolVar(&domainUseARecord, "use-a-record", false, "Use A DNS record")
	}

	domainsCmd.AddCommand(
		domainsListCmd, domainsGetCmd, domainsDeleteCmd,
		domainsCreateCmd, domainsUpdateCmd, domainsValidateCmd, domainsRevalidateCmd,
	)
}
