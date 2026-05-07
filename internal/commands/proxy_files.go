package commands

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/output"
)

var proxyFilesUpdateFromFile string

var proxyFilesCmd = &cobra.Command{
	Use:   "proxy-files",
	Short: "Manage container proxy file mappings",
}

var proxyFilesListCmd = &cobra.Command{
	Use:   "list <container>",
	Short: "List configured proxy files",
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
		data, err := c.ListProxyFiles(ctx(), id)
		if err != nil {
			return err
		}
		return output.PrintData("proxy-files.list", data, jsonFlag, jqFlag)
	},
}

var proxyFilesUpdateCmd = &cobra.Command{
	Use:   "update <container>",
	Short: "Replace the proxy file mappings (DESTRUCTIVE — requires --yes)",
	Long: `Replace the entire proxy-files configuration.

Body must satisfy ContainerProxyFileEditForm:
  {
    "containerProxyFiles": [
      {
        "originalFileUrl": "https://...",
        "customPath": "/path",
        "cacheMaxAge": 3600
      }
    ]
  }

cacheMaxAge enum: -1, 120, 300, 1200, 1800, 3600, 7200, 10800, 14400, 18000,
28800, 43200, 57600, 72000, 86400, 172800, 259200, 345600, 432000, 691200,
1382400, 2073600, 2592000, 5184000, 15552000, 31536000.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if err := confirmDestructive("PUT", "/api/v2/containers/"+id+"/proxy-files"); err != nil {
			return err
		}
		body, err := readBodyFile(proxyFilesUpdateFromFile)
		if err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.UpdateProxyFiles(ctx(), id, json.RawMessage(body))
		if err != nil {
			return err
		}
		return output.PrintData("proxy-files.update", data, jsonFlag, jqFlag)
	},
}

func init() {
	proxyFilesUpdateCmd.Flags().StringVar(&proxyFilesUpdateFromFile, "from-file", "", "Path to JSON body (ContainerProxyFileEditForm)")
	_ = proxyFilesUpdateCmd.MarkFlagRequired("from-file")

	proxyFilesCmd.AddCommand(proxyFilesListCmd, proxyFilesUpdateCmd)
}
