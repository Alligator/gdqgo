package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alligator/gdqgo/internal/statsfile"
	"github.com/alligator/gdqgo/internal/tracker"
	"github.com/alligator/gdqgo/internal/twitch"
	"github.com/alligator/gdqgo/internal/youtube"
	"github.com/spf13/cobra"
)

type fetchOpts struct {
	name               string
	typ                string
	trackerMararthonId int
	twitchUserId       int
	youtubeChannelId   string
	step               string
}

var fo fetchOpts
var errs []error

var fetchCmd = &cobra.Command{
	Use:          "fetch [file]",
	Short:        "fetch all the things",
	Long:         ``,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		errs = make([]error, 0)

		var sf statsfile.StatsFile
		sf, err := statsfile.Read(args[0])
		if err != nil {
			if os.IsNotExist(err) {
				sf = statsfile.New(fo.name, fo.typ)
			} else {
				return err
			}
		}

		step("fetch schedule", func() error {
			schedule, err := tracker.GetSchedule(fo.trackerMararthonId)
			if err != nil {
				return err
			}
			games := make([]statsfile.Game, 0, len(schedule))
			for _, g := range schedule {
				runners := make([]string, 0)
				for _, r := range g.Runners {
					runners = append(runners, r.Name)
				}

				games = append(games, statsfile.Game{
					Start:    float64(g.StartTime.UnixMilli()) / 1000,
					Name:     g.Name,
					Category: g.Category,
					Runners:  strings.Join(runners, ", "),
				})
			}

			sf.Games = games
			return nil
		})

		v := statsfile.Viewer{}
		now := time.Now().UTC()
		v.Time = float64(now.UnixMilli()) / 1000

		step("fetch donations", func() error {
			donations, err := tracker.GetDonations(fo.trackerMararthonId)
			if err != nil {
				return err
			}
			v.DonationTotal = &donations
			return nil
		})

		step("fetch twitch viewers", func() error {
			viewers, err := twitch.GetViewers(fo.twitchUserId)
			if err != nil {
				return err
			}
			i64 := int64(viewers)
			v.TwitchViewers = &i64
			return nil
		})

		step("fetch youtube viewers", func() error {
			viewers, err := youtube.GetViewers(fo.youtubeChannelId)
			if err != nil {
				return err
			}
			i64 := int64(0)
			if viewers.Live {
				i64 = int64(viewers.Viewers)
			}
			v.YoutubeViewers = &i64
			return nil
		})

		sf.Viewers = append(sf.Viewers, v)

		if err := statsfile.Write(args[0], sf); err != nil {
			errs = append(errs, err)
		}

		if len(errs) > 0 {
			return errors.Join(errs...)
		}

		return nil
	},
}

func step(name string, fn func() error) {
	defer func() {
		if r := recover(); r != nil {
			errs = append(errs, fmt.Errorf("[%s] panic: %s", name, r))
		}
	}()

	if fo.step != "" && name != fo.step {
		return
	}

	if err := fn(); err != nil {
		errs = append(errs, fmt.Errorf("[%s]: %w", name, err))
	}
}

func typeFlag(s string) error {
	switch s {
	case "gdq", "gdqx", "ff", "btb", "gdqueer":
		fo.typ = s
		return nil
	default:
		return fmt.Errorf("type must be one of 'gdq', 'gdqx', 'ff', 'gdqueer'")
	}
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	fetchCmd.Flags().StringVar(&fo.name, "name", "", "name")
	fetchCmd.Flags().Func("type", "one of 'gdq', 'gdqx', 'ff', 'btb' or 'gdqueer'", typeFlag)
	fetchCmd.Flags().IntVar(&fo.trackerMararthonId, "tracker-marathon-id", 0, "tracker marathon id")
	fetchCmd.Flags().IntVar(&fo.twitchUserId, "twitch-user-id", 0, "twitch user id")
	fetchCmd.Flags().StringVar(&fo.youtubeChannelId, "youtube-channel-id", "", "youtube channel id")
	fetchCmd.Flags().StringVar(&fo.step, "step", "", "only run this step")

	fetchCmd.MarkFlagRequired("name")
	fetchCmd.MarkFlagRequired("type")
	fetchCmd.MarkFlagRequired("tracker-marathon-id")
	fetchCmd.MarkFlagRequired("twitch-user-id")
	fetchCmd.MarkFlagRequired("youtube-channel-id")
}
