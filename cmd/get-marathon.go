package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/alligator/gdqgo/internal/tracker"
	"github.com/alligator/tbl"
	"github.com/spf13/cobra"
)

var getDatesCmd = &cobra.Command{
	Use:   "get-dates [id]",
	Short: "get the start and end date of a marathon",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		runs, err := tracker.GetSchedule(id)
		if err != nil {
			return err
		}

		t := tbl.NewTable()
		t.Style = tbl.StyleMinimal

		runsToPrint := make([]tracker.Run, 2)
		runsToPrint[0] = runs[0]
		runsToPrint[1] = runs[len(runs)-1]

		for _, run := range runsToPrint {
			t.NewRow()

			t.NewCol("Name")
			t.Print(run.Name)

			t.NewCol("Start")
			t.Printf("%s", run.StartTime.Format(time.RFC3339))

			t.NewCol("End")
			t.Printf("%s", run.EndTime.Format(time.RFC3339))
		}

		fmt.Println(t.String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getDatesCmd)
}
