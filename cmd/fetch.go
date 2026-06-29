package cmd

import (
	"fmt"

	"github.com/alligator/gdqgo/internal/fetch"
	"github.com/spf13/cobra"
)

var fo fetch.FetchOpts

var fetchCmd = &cobra.Command{
	Use:          "fetch [file]",
	Short:        "fetch all the things",
	Long:         ``,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		f := fetch.NewFetcher(fo)
		return f.DoFetch(args[0])
	},
}

func typeFlag(s string) error {
	switch s {
	case "gdq", "gdqx", "ff", "btb", "gdqueer":
		fo.Typ = s
		return nil
	default:
		return fmt.Errorf("type must be one of 'gdq', 'gdqx', 'ff', 'btb' or 'gdqueer'")
	}
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	fetchCmd.Flags().StringVar(&fo.Name, "name", "", "name")
	fetchCmd.Flags().Func("type", "one of 'gdq', 'gdqx', 'ff', 'btb' or 'gdqueer'", typeFlag)
	fetchCmd.Flags().IntVar(&fo.TrackerMararthonId, "tracker-marathon-id", 0, "tracker marathon id")
	fetchCmd.Flags().IntVar(&fo.TwitchUserId, "twitch-user-id", 0, "twitch user id")
	fetchCmd.Flags().StringVar(&fo.YoutubeChannelId, "youtube-channel-id", "", "youtube channel id")
	fetchCmd.Flags().StringVar(&fo.Step, "step", "", "only run this step")

	fetchCmd.MarkFlagRequired("name")
	fetchCmd.MarkFlagRequired("type")
	fetchCmd.MarkFlagRequired("tracker-marathon-id")
	fetchCmd.MarkFlagRequired("twitch-user-id")
	fetchCmd.MarkFlagRequired("youtube-channel-id")
}
