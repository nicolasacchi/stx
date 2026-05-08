package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/output"
)

var partnerCheckerCmd = &cobra.Command{
	Use:   "partner-checker",
	Short: "Stape Partner Tracking Checker",
	Long: `Run a tracking-checker report or query the monthly quota.

Each 'partner-checker create' consumes one quota unit; check 'limit' first
when running automation.`,
}

var (
	pcCreateSiteURL     string
	pcCreateCallbackURL string
)

var partnerCheckerCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a tracking checker report (DESTRUCTIVE — consumes monthly quota; requires --yes)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if pcCreateSiteURL == "" {
			return fmt.Errorf("--site-url required (full URL like https://example.com)")
		}
		if pcCreateCallbackURL == "" {
			return fmt.Errorf("--callback-url required (full URL)")
		}
		if err := confirmDestructive("POST", "/api/v2/partner-tracking-checker"); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.CreatePartnerChecker(ctx(), pcCreateSiteURL, pcCreateCallbackURL)
		if err != nil {
			return err
		}
		return output.PrintData("partner-checker.create", data, jsonFlag, jqFlag)
	},
}

var partnerCheckerLimitCmd = &cobra.Command{
	Use:   "limit",
	Short: "Get remaining monthly quota",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.GetPartnerCheckerLimit(ctx())
		if err != nil {
			return err
		}
		return output.PrintData("partner-checker.limit", data, jsonFlag, jqFlag)
	},
}

func init() {
	partnerCheckerCreateCmd.Flags().StringVar(&pcCreateSiteURL, "site-url", "", "Full site URL to check (required)")
	partnerCheckerCreateCmd.Flags().StringVar(&pcCreateCallbackURL, "callback-url", "", "Full callback URL for the report (required)")
	partnerCheckerCmd.AddCommand(partnerCheckerCreateCmd, partnerCheckerLimitCmd)
}
