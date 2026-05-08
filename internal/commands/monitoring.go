package commands

import (
	"encoding/json"
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

// --- Rules sub-group ---

var (
	rulesCreateFromFile string
	rulesUpdateFromFile string
	rulesSwitchOn       bool
	rulesSwitchOff      bool
)

var monRulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Manage monitoring rules (alert thresholds on log volume / events)",
}

var monRulesListCmd = &cobra.Command{
	Use:   "list <container>",
	Short: "List monitoring rules",
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
		data, err := c.ListRules(ctx(), id)
		if err != nil {
			return err
		}
		return output.PrintData("monitoring.rules.list", data, jsonFlag, jqFlag)
	},
}

var monRulesGetCmd = &cobra.Command{
	Use:   "get <container> <rule-id>",
	Short: "Get a single monitoring rule",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.GetRule(ctx(), args[0], args[1])
		if err != nil {
			return err
		}
		return output.PrintData("monitoring.rules.get", data, jsonFlag, jqFlag)
	},
}

var monRulesCreateCmd = &cobra.Command{
	Use:   "create <container>",
	Short: "Create a monitoring rule (DESTRUCTIVE — requires --yes)",
	Long: `Create a monitoring rule.

Body (ContainerMonitoringFormType) is too complex for flat flags — use --from-file.
Required fields: name, logType (Incoming|Outgoing), period (1|2|6|12|24|168),
comparisonType (lessThan|greaterThan|decreasesBy|increasesBy|percentDecreasesBy|percentIncreasesBy),
comparisonTarget (prevDay|sameDayPrevWeek|prevWeek), conditions (array).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		container := args[0]
		body, err := readBodyFile(rulesCreateFromFile)
		if err != nil {
			return err
		}
		if err := confirmDestructive("POST", "/api/v2/containers/"+container+"/monitoring"); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.CreateRule(ctx(), container, json.RawMessage(body))
		if err != nil {
			return err
		}
		return output.PrintData("monitoring.rules.create", data, jsonFlag, jqFlag)
	},
}

var monRulesUpdateCmd = &cobra.Command{
	Use:   "update <container> <rule-id>",
	Short: "Update a monitoring rule (full PUT replace) — requires --yes",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		container, rule := args[0], args[1]
		body, err := readBodyFile(rulesUpdateFromFile)
		if err != nil {
			return err
		}
		if err := confirmDestructive("PUT", "/api/v2/containers/"+container+"/monitoring/"+rule); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.UpdateRule(ctx(), container, rule, json.RawMessage(body))
		if err != nil {
			return err
		}
		return output.PrintData("monitoring.rules.update", data, jsonFlag, jqFlag)
	},
}

var monRulesDeleteCmd = &cobra.Command{
	Use:   "delete <container> <rule-id>",
	Short: "Delete a monitoring rule (DESTRUCTIVE — requires --yes)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		container, rule := args[0], args[1]
		if err := confirmDestructive("DELETE", "/api/v2/containers/"+container+"/monitoring/"+rule); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.DeleteRule(ctx(), container, rule)
		if err != nil {
			return err
		}
		return output.PrintData("monitoring.rules.delete", data, jsonFlag, jqFlag)
	},
}

var monRulesSwitchCmd = &cobra.Command{
	Use:   "switch <container> <rule-id>",
	Short: "Enable / disable a monitoring rule (--on/--off; --yes required)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		container, rule := args[0], args[1]
		enabled, err := resolveOnOff(rulesSwitchOn, rulesSwitchOff)
		if err != nil {
			return err
		}
		if err := confirmDestructive("PATCH", "/api/v2/containers/"+container+"/monitoring/"+rule); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.SwitchRule(ctx(), container, rule, enabled)
		if err != nil {
			return err
		}
		return output.PrintData("monitoring.rules.switch", data, jsonFlag, jqFlag)
	},
}

var monRulesResolveCmd = &cobra.Command{
	Use:   "resolve <container> <rule-id>",
	Short: "Mark a fired rule as resolved (--yes required)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		container, rule := args[0], args[1]
		if err := confirmDestructive("POST", "/api/v2/containers/"+container+"/monitoring/"+rule+"/resolve"); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.ResolveRule(ctx(), container, rule)
		if err != nil {
			return err
		}
		return output.PrintData("monitoring.rules.resolve", data, jsonFlag, jqFlag)
	},
}

// --- Emails sub-group ---

var (
	emailsAddEmail    string
	emailsDeleteEmail string
	emailsSwitchEmail string
	emailsSwitchOn    bool
	emailsSwitchOff   bool
)

var monEmailsCmd = &cobra.Command{
	Use:   "emails",
	Short: "Manage alert email addresses for monitoring rules",
}

var monEmailsListCmd = &cobra.Command{
	Use:   "list <container>",
	Short: "List configured alert emails",
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
		data, err := c.ListEmails(ctx(), id)
		if err != nil {
			return err
		}
		return output.PrintData("monitoring.emails.list", data, jsonFlag, jqFlag)
	},
}

var monEmailsAddCmd = &cobra.Command{
	Use:   "add <container>",
	Short: "Add an email to the alert recipients (--yes required)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		container := args[0]
		if emailsAddEmail == "" {
			return fmt.Errorf("--email required")
		}
		if err := confirmDestructive("POST", "/api/v2/containers/"+container+"/monitoring/emails"); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.AddEmail(ctx(), container, emailsAddEmail)
		if err != nil {
			return err
		}
		return output.PrintData("monitoring.emails.add", data, jsonFlag, jqFlag)
	},
}

var monEmailsDeleteCmd = &cobra.Command{
	Use:   "delete <container>",
	Short: "Remove an email from the alert recipients (--yes required)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		container := args[0]
		if emailsDeleteEmail == "" {
			return fmt.Errorf("--email required")
		}
		if err := confirmDestructive("DELETE", "/api/v2/containers/"+container+"/monitoring/emails/"+emailsDeleteEmail); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.DeleteEmail(ctx(), container, emailsDeleteEmail)
		if err != nil {
			return err
		}
		return output.PrintData("monitoring.emails.delete", data, jsonFlag, jqFlag)
	},
}

var monEmailsSwitchCmd = &cobra.Command{
	Use:   "switch <container>",
	Short: "Enable / disable a specific alert email (--on/--off; --yes required)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		container := args[0]
		if emailsSwitchEmail == "" {
			return fmt.Errorf("--email required")
		}
		enabled, err := resolveOnOff(emailsSwitchOn, emailsSwitchOff)
		if err != nil {
			return err
		}
		if err := confirmDestructive("PATCH", "/api/v2/containers/"+container+"/monitoring/switch-email"); err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.SwitchEmail(ctx(), container, emailsSwitchEmail, enabled)
		if err != nil {
			return err
		}
		return output.PrintData("monitoring.emails.switch", data, jsonFlag, jqFlag)
	},
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

	// Rules
	monRulesCreateCmd.Flags().StringVar(&rulesCreateFromFile, "from-file", "", "Path to JSON body (ContainerMonitoringFormType)")
	_ = monRulesCreateCmd.MarkFlagRequired("from-file")
	monRulesUpdateCmd.Flags().StringVar(&rulesUpdateFromFile, "from-file", "", "Path to JSON body (ContainerMonitoringFormType)")
	_ = monRulesUpdateCmd.MarkFlagRequired("from-file")
	monRulesSwitchCmd.Flags().BoolVar(&rulesSwitchOn, "on", false, "Enable the rule")
	monRulesSwitchCmd.Flags().BoolVar(&rulesSwitchOff, "off", false, "Disable the rule")
	monRulesCmd.AddCommand(
		monRulesListCmd, monRulesGetCmd,
		monRulesCreateCmd, monRulesUpdateCmd, monRulesDeleteCmd,
		monRulesSwitchCmd, monRulesResolveCmd,
	)

	// Emails
	monEmailsAddCmd.Flags().StringVar(&emailsAddEmail, "email", "", "Email address to add (required)")
	_ = monEmailsAddCmd.MarkFlagRequired("email")
	monEmailsDeleteCmd.Flags().StringVar(&emailsDeleteEmail, "email", "", "Email address to delete (required)")
	_ = monEmailsDeleteCmd.MarkFlagRequired("email")
	monEmailsSwitchCmd.Flags().StringVar(&emailsSwitchEmail, "email", "", "Email address (required)")
	monEmailsSwitchCmd.Flags().BoolVar(&emailsSwitchOn, "on", false, "Enable")
	monEmailsSwitchCmd.Flags().BoolVar(&emailsSwitchOff, "off", false, "Disable")
	_ = monEmailsSwitchCmd.MarkFlagRequired("email")
	monEmailsCmd.AddCommand(monEmailsListCmd, monEmailsAddCmd, monEmailsDeleteCmd, monEmailsSwitchCmd)

	monitoringCmd.AddCommand(logsCmd, monRulesCmd, monEmailsCmd)
}
