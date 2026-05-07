package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/client"
	"github.com/nicolasacchi/stx/internal/output"
	"github.com/nicolasacchi/stx/internal/timeparse"
)

var (
	logsFromFlag      string
	logsToFlag        string
	logsSinceFlag     string
	logsPlatformFlag  string
	logsEventTypeFlag string
	traceIDFlag       string
	traceDateFlag     string
)

var monitoringCmd = &cobra.Command{
	Use:   "monitoring",
	Short: "Monitoring rules, emails, and outgoing logs",
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Outgoing request logs (replaces manual log.csv panel download)",
}

var logsAggregatedCmd = &cobra.Command{
	Use:   "aggregated <container>",
	Short: "Aggregated outgoing log counts",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLogsRange(args, "monitoring.logs.aggregated", func(c *client.Client, id string, q client.LogsQuery) ([]byte, error) {
			return c.ListLogsAggregated(ctx(), id, q)
		})
	},
}

var logsDetailedCmd = &cobra.Command{
	Use:   "detailed <container>",
	Short: "Detailed outgoing log entries (per request)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLogsRange(args, "monitoring.logs.detailed", func(c *client.Client, id string, q client.LogsQuery) ([]byte, error) {
			return c.ListLogsDetailed(ctx(), id, q)
		})
	},
}

var logsTraceCmd = &cobra.Command{
	Use:   "trace <container>",
	Short: "Single trace by ID",
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
		if traceIDFlag == "" {
			return fmt.Errorf("--trace-id required")
		}
		// API requires `date` even though the spec marks it optional.
		// Default to "now" so the user doesn't have to.
		dateInput := traceDateFlag
		if dateInput == "" {
			dateInput = "now"
		}
		date, err := timeparse.Parse(dateInput)
		if err != nil {
			return fmt.Errorf("--date: %w", err)
		}
		data, err := c.GetLogTrace(ctx(), id, traceIDFlag, date)
		if err != nil {
			return err
		}
		return output.PrintData("monitoring.logs.trace", data, jsonFlag, jqFlag)
	},
}

func runLogsRange(args []string, cmdKey string, call func(*client.Client, string, client.LogsQuery) ([]byte, error)) error {
	c, creds, err := getClient()
	if err != nil {
		return err
	}
	id, err := resolveContainer(args, creds)
	if err != nil {
		return err
	}
	q, err := resolveLogsRange()
	if err != nil {
		return err
	}
	data, err := call(c, id, q)
	if err != nil {
		return err
	}
	return output.PrintData(cmdKey, data, jsonFlag, jqFlag)
}

func resolveLogsRange() (client.LogsQuery, error) {
	q := client.LogsQuery{Platform: logsPlatformFlag, EventType: logsEventTypeFlag}
	if logsSinceFlag != "" && (logsFromFlag != "" || logsToFlag != "") {
		return q, fmt.Errorf("--since cannot be combined with --from / --to")
	}
	if logsSinceFlag != "" {
		start, err := timeparse.Parse(logsSinceFlag)
		if err != nil {
			return q, fmt.Errorf("--since: %w", err)
		}
		q.Start = start
		q.End, _ = timeparse.Parse("now")
		return q, nil
	}
	if logsFromFlag != "" {
		v, err := timeparse.Parse(logsFromFlag)
		if err != nil {
			return q, fmt.Errorf("--from: %w", err)
		}
		q.Start = v
	}
	if logsToFlag != "" {
		v, err := timeparse.Parse(logsToFlag)
		if err != nil {
			return q, fmt.Errorf("--to: %w", err)
		}
		q.End = v
	}
	return q, nil
}

func init() {
	for _, c := range []*cobra.Command{logsAggregatedCmd, logsDetailedCmd} {
		c.Flags().StringVar(&logsFromFlag, "from", "", "start time (e.g. 1h ago, RFC3339, epoch)")
		c.Flags().StringVar(&logsToFlag, "to", "", "end time (default: now when --from set)")
		c.Flags().StringVar(&logsSinceFlag, "since", "", "shorthand: start=now-<dur>, end=now (e.g. 5m, 1h)")
		c.Flags().StringVar(&logsPlatformFlag, "platform", "", "filter by platform (e.g. Facebook)")
		c.Flags().StringVar(&logsEventTypeFlag, "event-type", "", "filter by event type (e.g. PageView)")
	}

	logsTraceCmd.Flags().StringVar(&traceIDFlag, "trace-id", "", "trace UUID (required)")
	logsTraceCmd.Flags().StringVar(&traceDateFlag, "date", "", "date for the trace (default: now)")
	_ = logsTraceCmd.MarkFlagRequired("trace-id")

	logsCmd.AddCommand(logsAggregatedCmd, logsDetailedCmd, logsTraceCmd)
	monitoringCmd.AddCommand(logsCmd)
}
