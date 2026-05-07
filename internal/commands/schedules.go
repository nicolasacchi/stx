package commands

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/nicolasacchi/stx/internal/output"
)

var schedulesUpdateFromFile string

var schedulesCmd = &cobra.Command{
	Use:   "schedules",
	Short: "Manage container scheduled tasks",
}

var schedulesListCmd = &cobra.Command{
	Use:   "list <container>",
	Short: "List configured schedules",
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
		data, err := c.ListSchedules(ctx(), id)
		if err != nil {
			return err
		}
		return output.PrintData("schedules.list", data, jsonFlag, jqFlag)
	},
}

var schedulesUpdateCmd = &cobra.Command{
	Use:   "update <container>",
	Short: "Replace the schedules (DESTRUCTIVE — requires --yes)",
	Long: `Replace the entire schedules configuration.

Body must satisfy ContainerScheduleEditForm:
  {
    "containerSchedules": [
      {
        "domain": {"identifier": "<domain-uuid>"},
        "path": "/some/path",
        "frequencyType": "onceADay",
        "hour": 3,
        "minute": 0
      }
    ]
  }

frequencyType enum: onceADay, everyHour.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if err := confirmDestructive("PUT", "/api/v2/containers/"+id+"/schedules"); err != nil {
			return err
		}
		body, err := readBodyFile(schedulesUpdateFromFile)
		if err != nil {
			return err
		}
		c, _, err := getClient()
		if err != nil {
			return err
		}
		data, err := c.UpdateSchedules(ctx(), id, json.RawMessage(body))
		if err != nil {
			return err
		}
		return output.PrintData("schedules.update", data, jsonFlag, jqFlag)
	},
}

func init() {
	schedulesUpdateCmd.Flags().StringVar(&schedulesUpdateFromFile, "from-file", "", "Path to JSON body (ContainerScheduleEditForm)")
	_ = schedulesUpdateCmd.MarkFlagRequired("from-file")

	schedulesCmd.AddCommand(schedulesListCmd, schedulesUpdateCmd)
}
