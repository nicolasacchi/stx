package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/nicolasacchi/stx/internal/client"
	"github.com/nicolasacchi/stx/internal/timeparse"
)

type fetchResult struct {
	data []byte
	err  error
}

var overviewCmd = &cobra.Command{
	Use:   "overview <container>",
	Short: "Parallel-fetch dashboard for a container",
	Long: `Fan-out 5 read calls in parallel and render a multi-section summary:
container details, custom-loader / sGTM state, domains, monitoring config,
recent traffic. Useful for daily ops + ` + "`/health-check`" + ` integration.

Use --json to get a single combined JSON object instead of the dashboard.`,
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

		var (
			container, domains, rules, emails, logs fetchResult
			wg                                      sync.WaitGroup
		)
		wg.Add(5)
		go func() {
			defer wg.Done()
			container.data, container.err = c.GetContainer(ctx(), id)
		}()
		go func() {
			defer wg.Done()
			domains.data, domains.err = c.ListDomains(ctx(), id)
		}()
		go func() {
			defer wg.Done()
			rules.data, rules.err = c.ListRules(ctx(), id)
		}()
		go func() {
			defer wg.Done()
			emails.data, emails.err = c.ListEmails(ctx(), id)
		}()
		go func() {
			defer wg.Done()
			now, _ := timeparse.Parse("now")
			start, _ := timeparse.Parse("1h")
			logs.data, logs.err = c.ListLogsAggregated(ctx(), id, client.LogsQuery{Start: start, End: now})
		}()
		wg.Wait()

		if jsonFlag || !isStdoutTTY() {
			combined := map[string]json.RawMessage{}
			addResult := func(k string, r fetchResult) {
				if r.err != nil {
					combined[k] = json.RawMessage(fmt.Sprintf(`{"error": %q}`, r.err.Error()))
				} else {
					combined[k] = json.RawMessage(r.data)
				}
			}
			addResult("container", container)
			addResult("domains", domains)
			addResult("rules", rules)
			addResult("emails", emails)
			addResult("logs_1h", logs)
			out, _ := json.MarshalIndent(combined, "", "  ")
			if jqFlag != "" {
				res := gjson.GetBytes(out, jqFlag)
				if !res.Exists() {
					fmt.Println("null")
				} else {
					fmt.Println(res.Raw)
				}
				return nil
			}
			fmt.Println(string(out))
			return nil
		}

		renderDashboard(id, container, domains, rules, emails, logs)
		return nil
	},
}

func renderDashboard(id string, container, domains, rules, emails, logs fetchResult) {
	fmt.Printf("\nContainer %s\n", id)
	if container.err != nil {
		fmt.Printf("  (error: %s)\n", container.err)
	} else {
		body := container.data
		print := func(label, path string, transform func(gjson.Result) string) {
			r := gjson.GetBytes(body, "body."+path)
			val := ""
			if r.Exists() {
				if transform != nil {
					val = transform(r)
				} else {
					val = r.String()
				}
			}
			if val != "" {
				fmt.Printf("  %-13s%s\n", label+":", val)
			}
		}
		print("Name", "name", nil)
		print("Status", "status.label", nil)
		print("Plan", "subscription.plan.label", func(r gjson.Result) string {
			limit := gjson.GetBytes(body, "body.subscription.plan.features.requestLimit")
			price := gjson.GetBytes(body, "body.subscription.plan.price.month")
			s := r.String()
			if price.Exists() {
				s += fmt.Sprintf(" (€%v/mo", price.Num)
				if limit.Exists() {
					s += fmt.Sprintf(", %s req/mo", humanInt(int64(limit.Num)))
				}
				s += ")"
			}
			return s
		})
		usage := gjson.GetBytes(body, "body.subscription.usageCount")
		limit := gjson.GetBytes(body, "body.subscription.plan.features.requestLimit")
		if usage.Exists() {
			pct := ""
			if limit.Exists() && limit.Num > 0 {
				pct = fmt.Sprintf(" (%.2f%% of cap)", 100*usage.Num/limit.Num)
			}
			fmt.Printf("  %-13s%s events%s\n", "Usage:", humanInt(int64(usage.Num)), pct)
		}
		print("Zone", "zone.label", func(r gjson.Result) string {
			z := gjson.GetBytes(body, "body.zone.type")
			return r.String() + " — " + z.String()
		})
		print("Created", "createdAt", func(r gjson.Result) string {
			return time.Unix(int64(r.Num), 0).Format("2006-01-02")
		})
	}

	// Custom loader summary (from container.body.powerUps + container.body.sGtmContainerId)
	fmt.Println()
	if container.err == nil {
		cl := gjson.GetBytes(container.data, "body.powerUps.customLoader")
		state := "DISABLED"
		if cl.Bool() {
			state = "ENABLED"
		}
		fmt.Printf("Custom Loader  %s\n", state)
		// Pick the first non-stape domain as the "loader domain"
		dom := gjson.GetBytes(container.data, "body.zone.defaultDomain")
		if domains.err == nil {
			first := gjson.GetBytes(domains.data, "body.items.0.name")
			if first.Exists() {
				fmt.Printf("  %-13s%s", "Domain:", first.String())
				st := gjson.GetBytes(domains.data, "body.items.0.status.label")
				if st.Exists() {
					fmt.Printf(" (%s)", st.String())
				}
				fmt.Println()
			}
		}
		_ = dom
		sgtm := gjson.GetBytes(container.data, "body.sGtmContainerId")
		if sgtm.Exists() {
			fmt.Printf("  %-13s%s\n", "sGTM ID:", sgtm.String())
		}
	}

	// Domains
	fmt.Println()
	if domains.err != nil {
		fmt.Printf("Domains  (error: %s)\n", domains.err)
	} else {
		total := gjson.GetBytes(domains.data, "body.total")
		fmt.Printf("Domains (%v)\n", total.Int())
		gjson.GetBytes(domains.data, "body.items").ForEach(func(_, v gjson.Result) bool {
			name := v.Get("name").String()
			cdn := v.Get("cdnType").String()
			st := v.Get("status.label").String()
			fmt.Printf("  %-32s  %-7s  %s\n", name, cdn, st)
			return true
		})
	}

	// Monitoring (rules + emails)
	fmt.Println()
	fmt.Println("Monitoring")
	if rules.err != nil {
		fmt.Printf("  Rules:       (error: %s)\n", rules.err)
	} else {
		t := gjson.GetBytes(rules.data, "body.total")
		fmt.Printf("  %-13s%v configured\n", "Rules:", t.Int())
	}
	if emails.err != nil {
		fmt.Printf("  Emails:      (error: %s)\n", emails.err)
	} else {
		entries := []string{}
		gjson.GetBytes(emails.data, "body.items").ForEach(func(_, v gjson.Result) bool {
			s := v.Get("email").String()
			if v.Get("isDefault").Bool() {
				s += " ★"
			}
			if !v.Get("isEnabled").Bool() {
				s += " (disabled)"
			}
			entries = append(entries, s)
			return true
		})
		t := gjson.GetBytes(emails.data, "body.total")
		fmt.Printf("  %-13s%v", "Emails:", t.Int())
		if len(entries) > 0 {
			fmt.Printf(" (%s)", joinComma(entries))
		}
		fmt.Println()
	}

	// Recent traffic (last hour)
	fmt.Println()
	fmt.Println("Recent traffic (last hour)")
	if logs.err != nil {
		fmt.Printf("  (error: %s)\n", logs.err)
	} else {
		body := gjson.GetBytes(logs.data, "body")
		if !body.Exists() || body.IsArray() && body.Get("#").Int() == 0 {
			fmt.Println("  (no aggregated data)")
		} else {
			// body is an array of {date, count, ...}; sum counts.
			var total int64
			body.ForEach(func(_, v gjson.Result) bool {
				total += v.Get("count").Int()
				return true
			})
			fmt.Printf("  %-13s%s events\n", "Total:", humanInt(total))
		}
	}

	// Analytics module state (from container.body.analyticsEnabledAt)
	fmt.Println()
	fmt.Println("Analytics")
	if container.err == nil {
		enabled := gjson.GetBytes(container.data, "body.analyticsEnabledAt")
		if enabled.Type == gjson.Null || !enabled.Exists() {
			fmt.Println("  Module:      not enabled")
		} else {
			fmt.Printf("  Module:      enabled at %s\n", time.Unix(enabled.Int(), 0).Format("2006-01-02"))
		}
	}
	fmt.Println()
}

func humanInt(n int64) string {
	// 1234567 -> "1,234,567"
	s := fmt.Sprintf("%d", n)
	if n < 1000 {
		return s
	}
	out := ""
	for i, ch := range reverse(s) {
		if i > 0 && i%3 == 0 {
			out = "," + out
		}
		out = string(ch) + out
	}
	return out
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func joinComma(items []string) string {
	out := ""
	for i, s := range items {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}

func isStdoutTTY() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func init() {
	rootCmd.AddCommand(overviewCmd)
}
