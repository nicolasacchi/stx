package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/client"
	"github.com/nicolasacchi/stx/internal/output"
	"github.com/nicolasacchi/stx/internal/timeparse"
)

var (
	analyticsFromFlag  string
	analyticsToFlag    string
	analyticsSinceFlag string
)

var analyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "Container analytics (subscription usage by client / browser)",
}

var analyticsEnableOff bool

var analyticsEnableCmd = &cobra.Command{
	Use:   "enable <container>",
	Short: "Enable (or disable with --off) the analytics module",
	Long: `PATCH the analytics module on. Required to populate
'analytics browsers' / 'analytics clients' data.

Pass --off to disable. Requires --yes (DESTRUCTIVE).`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, creds, err := getClient()
		if err != nil {
			return err
		}
		id, err := resolveContainer(args, creds)
		if err != nil {
			return err
		}
		enabled := !analyticsEnableOff
		if err := confirmDestructive("PATCH", "/api/v2/containers/"+id+"/analytics-enable"); err != nil {
			return err
		}
		data, err := c.EnableAnalytics(ctx(), id, enabled)
		if err != nil {
			return err
		}
		return output.PrintData("analytics.enable", data, jsonFlag, jqFlag)
	},
}

var analyticsInfoCmd = &cobra.Command{
	Use:   "info <container>",
	Short: "Get analytics info for a container",
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
		data, err := c.GetAnalyticsInfo(ctx(), id)
		if err != nil {
			return err
		}
		return output.PrintData("analytics.info", data, jsonFlag, jqFlag)
	},
}

var analyticsBrowsersCmd = &cobra.Command{
	Use:   "browsers <container>",
	Short: "Browser breakdown of subscription usage (start/end as Unix seconds)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAnalyticsRange(args, "analytics.browsers", func(c *client.Client, id string, q client.AnalyticsQuery) ([]byte, error) {
			return c.GetAnalyticsBrowsers(ctx(), id, q)
		})
	},
}

var analyticsClientsCmd = &cobra.Command{
	Use:   "clients <container>",
	Short: "Client breakdown of subscription usage (start/end as Unix seconds)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAnalyticsRange(args, "analytics.clients", func(c *client.Client, id string, q client.AnalyticsQuery) ([]byte, error) {
			return c.GetAnalyticsClients(ctx(), id, q)
		})
	},
}

func runAnalyticsRange(args []string, cmdKey string, call func(*client.Client, string, client.AnalyticsQuery) ([]byte, error)) error {
	c, creds, err := getClient()
	if err != nil {
		return err
	}
	id, err := resolveContainer(args, creds)
	if err != nil {
		return err
	}
	q, err := resolveAnalyticsRange()
	if err != nil {
		return err
	}
	data, err := call(c, id, q)
	if err != nil {
		return err
	}
	return output.PrintData(cmdKey, data, jsonFlag, jqFlag)
}

// resolveAnalyticsRange resolves --from/--to or --since into start/end (Unix seconds).
// --since is mutually exclusive with --from/--to.
func resolveAnalyticsRange() (client.AnalyticsQuery, error) {
	q := client.AnalyticsQuery{}
	if analyticsSinceFlag != "" && (analyticsFromFlag != "" || analyticsToFlag != "") {
		return q, fmt.Errorf("--since cannot be combined with --from / --to")
	}
	if analyticsSinceFlag != "" {
		start, err := timeparse.Parse(analyticsSinceFlag)
		if err != nil {
			return q, fmt.Errorf("--since: %w", err)
		}
		q.Start = start
		q.End, _ = timeparse.Parse("now")
		return q, nil
	}
	if analyticsFromFlag != "" {
		v, err := timeparse.Parse(analyticsFromFlag)
		if err != nil {
			return q, fmt.Errorf("--from: %w", err)
		}
		q.Start = v
	}
	if analyticsToFlag != "" {
		v, err := timeparse.Parse(analyticsToFlag)
		if err != nil {
			return q, fmt.Errorf("--to: %w", err)
		}
		q.End = v
	}
	return q, nil
}

func init() {
	for _, c := range []*cobra.Command{analyticsBrowsersCmd, analyticsClientsCmd} {
		c.Flags().StringVar(&analyticsFromFlag, "from", "", "start time (e.g. 1h ago, RFC3339, epoch)")
		c.Flags().StringVar(&analyticsToFlag, "to", "", "end time (default: now when --from set)")
		c.Flags().StringVar(&analyticsSinceFlag, "since", "", "shorthand: start=now-<dur>, end=now (e.g. 7d)")
	}
	analyticsEnableCmd.Flags().BoolVar(&analyticsEnableOff, "off", false, "Disable analytics instead of enabling")
	analyticsCmd.AddCommand(analyticsInfoCmd, analyticsBrowsersCmd, analyticsClientsCmd, analyticsEnableCmd)
}
