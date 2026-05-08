package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/output"
)

var powerUpsCmd = &cobra.Command{
	Use:   "power-ups",
	Short: "Inspect and toggle container power-ups",
	Long: `Container power-ups: ad-blocker, anonymizer, cookie-keeper, custom-loader, …

20 typed PATCH subcommands (Tier A — toggle only, Tier B — toggle + --options-file)
plus the ` + "`get`" + ` dispatch.`,
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

// registerPowerUpToggle constructs and registers a typed PATCH subcommand
// for one power-up type. hasOptions=true adds the --options-file flag;
// optionsRequired=true makes that file mandatory when --on is used.
func registerPowerUpToggle(parent *cobra.Command, name string, hasOptions, optionsRequired bool, aliases ...string) {
	var (
		on             bool
		off            bool
		optionsFile    string
	)
	short := "Toggle power-up: " + name
	if hasOptions {
		short += " (Tier B — accepts --options-file)"
	} else {
		short += " (Tier A — toggle only)"
	}
	long := short + "\n\nDESTRUCTIVE — requires --yes."
	if hasOptions && optionsRequired {
		long += "\n\n--options-file is REQUIRED when --on is used (spec marks `options` required for this power-up)."
	}
	if hasOptions && name == "service-account" {
		long += "\n\nNote: for service-account, --options-file contents are sent as a raw string\n(Google service-account credentials JSON blob), NOT parsed as JSON."
	}

	cmd := &cobra.Command{
		Use:     name + " <container>",
		Short:   short,
		Long:    long,
		Aliases: aliases,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, creds, err := getClient()
			if err != nil {
				return err
			}
			id, err := resolveContainer(args, creds)
			if err != nil {
				return err
			}
			isActive, err := resolveOnOff(on, off)
			if err != nil {
				return err
			}
			body := map[string]any{"isActive": isActive}
			if hasOptions {
				if optionsFile == "" && isActive && optionsRequired {
					return fmt.Errorf("--options-file required when enabling %q (spec marks options required)", name)
				}
				if optionsFile != "" {
					raw, err := os.ReadFile(optionsFile)
					if err != nil {
						return fmt.Errorf("read %s: %w", optionsFile, err)
					}
					if name == "service-account" {
						// options is a raw string (credentials JSON blob) not a parsed object
						body["options"] = string(raw)
					} else {
						var opts any
						if err := json.Unmarshal(raw, &opts); err != nil {
							return fmt.Errorf("%s: invalid JSON: %w", optionsFile, err)
						}
						body["options"] = opts
					}
				}
			}
			if err := confirmDestructive("PATCH", "/api/v2/containers/"+id+"/power-ups/"+name); err != nil {
				return err
			}
			data, err := c.SwitchPowerUp(ctx(), id, name, body)
			if err != nil {
				return err
			}
			return output.PrintData("power-ups."+name, data, jsonFlag, jqFlag)
		},
	}
	cmd.Flags().BoolVar(&on, "on", false, "Enable the power-up")
	cmd.Flags().BoolVar(&off, "off", false, "Disable the power-up")
	if hasOptions {
		desc := "JSON options body"
		if optionsRequired {
			desc += " (REQUIRED when enabling)"
		}
		cmd.Flags().StringVar(&optionsFile, "options-file", "", desc)
	}
	parent.AddCommand(cmd)
}

// resolveOnOff returns true if --on, false if --off, and errors on neither/both.
func resolveOnOff(on, off bool) (bool, error) {
	switch {
	case on && off:
		return false, fmt.Errorf("--on and --off are mutually exclusive")
	case on:
		return true, nil
	case off:
		return false, nil
	default:
		return false, fmt.Errorf("--on or --off required")
	}
}

func init() {
	powerUpsCmd.AddCommand(powerUpsGetCmd)

	// Tier A — simple toggles (8)
	for _, name := range []string{
		"ad-blocker",
		"bot-index",
		"custom-loader",
		"geo-headers",
		"request-delay",
		"user-agent-headers",
		"user-id",
		"xml-to-json",
	} {
		registerPowerUpToggle(powerUpsCmd, name, false, false)
	}

	// Tier B — toggle + options (12)
	type tierB struct {
		name             string
		optionsRequired  bool
		aliases          []string
	}
	for _, t := range []tierB{
		{"anonymizer", true, nil},
		{"block-request-by-ip", true, nil},
		{"bot-detection", true, nil},
		{"click-id-restorer", false, nil},
		{"cookie-keeper", false, nil},
		{"dedicated-ip", true, nil},
		{"enricher", false, nil},
		{"preview-header-config", false, []string{"header-config"}},
		{"product-feed", false, nil},
		{"proxy-files", true, nil},
		{"schedule", true, nil},
		{"service-account", false, nil},
	} {
		registerPowerUpToggle(powerUpsCmd, t.name, true, t.optionsRequired, t.aliases...)
	}

	// Suppress unused-import errors when strings is only used in get cmd above
	_ = strings.ToLower
}
