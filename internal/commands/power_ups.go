package commands

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/output"
)

var powerUpsCmd = &cobra.Command{
	Use:   "power-ups",
	Short: "Inspect and manage container power-ups",
	Long: `Container power-ups: ad-blocker, anonymizer, cookie-keeper, custom-loader, …

20 PATCH subcommands for toggling/configuring power-ups land in M4.
For now this group exposes only ` + "`get`" + ` for inspection.`,
}

var powerUpsGetCmd = &cobra.Command{
	Use:   "get <container> <type>",
	Short: "Get the state of a single power-up",
	Long: `Read the configuration + active state of one power-up type.

Type names match the spec exactly (kebab-case): ad-blocker, anonymizer,
block-request-by-ip, bot-detection, bot-index, click-id-restorer,
cookie-keeper, custom-loader, dedicated-ip, enricher, geo-headers,
preview-header-config, product-feed, proxy-files, request-delay,
schedule, service-account, user-agent-headers, user-id, xml-to-json.

CLI alias: header-config -> preview-header-config.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		container, kind := args[0], args[1]
		// Brief alias accepted by CLI
		if kind == "header-config" {
			kind = "preview-header-config"
		}
		// Normalise to kebab-case (be permissive about underscores/spaces)
		kind = strings.ReplaceAll(kind, "_", "-")
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.GetPowerUp(ctx(), container, kind)
		if err != nil {
			return err
		}
		return output.PrintData("power-ups.get", data, jsonFlag, jqFlag)
	},
}

func init() {
	powerUpsCmd.AddCommand(powerUpsGetCmd)
}
