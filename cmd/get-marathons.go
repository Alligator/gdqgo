package cmd

import (
	"fmt"
	"time"

	"github.com/alligator/gdqgo/internal/tracker"
	"github.com/alligator/tbl"
	"github.com/spf13/cobra"
)

var getMarathonsCmd = &cobra.Command{
	Use:   "get-marathons",
	Short: "get the latest 10 marathons",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		events, err := tracker.GetEvents()
		if err != nil {
			return err
		}

		t := tbl.NewTable()
		t.Style = tbl.StyleMinimal

		for _, event := range events[:10] {
			t.NewRow()

			t.NewCol("Id")
			t.Printf("%d", event.Id)

			t.NewCol("Start")
			t.Printf("%s", event.DateTime.Format(time.RFC3339))

			t.NewCol("Name")
			t.Print(event.Name)

			t.NewCol("Short")
			t.Print(event.Short)
		}

		fmt.Println(t.String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getMarathonsCmd)
}
