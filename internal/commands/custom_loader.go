package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/output"
)

var (
	clWebGtmID            string
	clDomain              string
	clSource              string
	clDataLayerObjectName string
	clUserIdentifierType  string
	clUserIdentifierValue string
	clSameOriginPath      string
	clFromFile            string
)

var customLoaderCmd = &cobra.Command{
	Use:   "custom-loader",
	Short: "Generate custom loader scripts (singular /container/ path)",
}

var customLoaderGenerateCmd = &cobra.Command{
	Use:   "generate <container>",
	Short: "Generate the custom loader script for a container",
	Long: `Generate the custom loader JS snippet to embed on a website.

Body fields:
  --web-gtm-id            client-side GTM container ID (e.g. GTM-XXXX) — REQUIRED in practice
  --domain                first-party domain serving the loader     — REQUIRED in practice
  --source                CMS hint: wordpress|magento|shopify|bigCommerce|prestaShop|wix|salla|odoo|drupal|other  — REQUIRED in practice
  --data-layer-object     dataLayer var name (default: "dataLayer")
  --user-identifier-type  cookie|cssSelector|localStorage|jsVariable|stapeUserId
  --user-identifier-value the cookie/selector/etc.
  --same-origin-path      proxy path for same-origin first-party fetches

The spec marks all fields optional, but the API rejects with 400 unless
webGtmId / domain / source are present. Keep that in mind.

Or pass the entire body via --from-file <path> (JSON).`,
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
		var body any
		if clFromFile != "" {
			b, err := readBodyFile(clFromFile)
			if err != nil {
				return err
			}
			body = json.RawMessage(b)
		} else {
			form := map[string]any{}
			setNonEmpty(form, "webGtmId", clWebGtmID)
			setNonEmpty(form, "domain", clDomain)
			setNonEmpty(form, "source", clSource)
			setNonEmpty(form, "dataLayerObjectName", clDataLayerObjectName)
			setNonEmpty(form, "userIdentifierType", clUserIdentifierType)
			setNonEmpty(form, "userIdentifierValue", clUserIdentifierValue)
			setNonEmpty(form, "sameOriginPath", clSameOriginPath)
			body = form
		}
		data, err := c.GenerateCustomLoader(ctx(), id, body)
		if err != nil {
			return err
		}
		return output.PrintData("custom-loader.generate", data, jsonFlag, jqFlag)
	},
}

func setNonEmpty(m map[string]any, k, v string) {
	if v != "" {
		m[k] = v
	}
}

func init() {
	g := customLoaderGenerateCmd.Flags()
	g.StringVar(&clWebGtmID, "web-gtm-id", "", "Client-side GTM container ID")
	g.StringVar(&clDomain, "domain", "", "First-party domain")
	g.StringVar(&clSource, "source", "", "CMS hint (wordpress|magento|shopify|bigCommerce|prestaShop|wix|salla|odoo|drupal|other)")
	g.StringVar(&clDataLayerObjectName, "data-layer-object", "", "dataLayer object name (default: dataLayer)")
	g.StringVar(&clUserIdentifierType, "user-identifier-type", "", "cookie|cssSelector|localStorage|jsVariable|stapeUserId")
	g.StringVar(&clUserIdentifierValue, "user-identifier-value", "", "Identifier value (cookie name, CSS selector, etc.)")
	g.StringVar(&clSameOriginPath, "same-origin-path", "", "Same-origin proxy path")
	g.StringVar(&clFromFile, "from-file", "", "JSON body file (overrides individual flags)")

	customLoaderCmd.AddCommand(customLoaderGenerateCmd)
	_ = fmt.Sprintf // keep imports tidy
}
