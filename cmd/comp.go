package cmd

import (
	"cmp"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/alligator/gdqgo/internal/statsfile"
	"github.com/spf13/cobra"
)

type compOpts struct {
	name  string
	globs []string
	days  int
}

type meta struct {
	MaxDonations   float64 `json:"max_donations"`
	MaxViewers     int64   `json:"max_viewers"`
	MaxViewersTs   float64 `json:"max_viewers_ts"`
	MaxViewersGame string  `json:"max_viewers_game"`
}

type compFile struct {
	Name      string       `json:"name"`
	Marathons []string     `json:"marathons"`
	Ts        []float64    `json:"ts"`
	Viewers   [][]*int64   `json:"viewers"`
	ViewersYt [][]*int64   `json:"viewers_yt"`
	Donations [][]*float64 `json:"donations"`
	Meta      []meta       `json:"meta"`
}

var co compOpts

var compCmd = &cobra.Command{
	Use:   "comp [output_file]",
	Short: "generate comparison JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		files, err := expand(co.globs)
		if err != nil {
			return err
		}

		if len(files) == 0 {
			return fmt.Errorf("no files found!")
		}

		sort(files)

		marathons := make([]statsfile.StatsFile, 0)
		marathonNames := make([]string, 0)

		for _, f := range files {
			if strings.HasSuffix(f, "sgdq14.json") {
				break
			}

			fmt.Printf("reading %s\n", f)
			m, err := statsfile.Read(f)
			if err != nil {
				return fmt.Errorf("error reading %s %w", f, err)
			}

			if strings.HasSuffix(f, "sgdq17.json") {
				// sgdq 17 hack
				// this has an extra day's worth of data at the start
				// so i chop it off
				m.Viewers = m.Viewers[1200:]
			}

			marathons = append(marathons, m)
			base := filepath.Base(f)
			base = strings.Split(base, ".")[0]

			marathonNames = append(marathonNames, base)
		}

		// find earliest start time
		earliestStart := slices.MinFunc(marathons, func(a statsfile.StatsFile, b statsfile.StatsFile) int {
			return cmp.Compare(a.Viewers[0].Time, b.Viewers[0].Time)
		}).Viewers[0].Time

		// find 3pm on that day - this is the root timestamp
		rootTs := threePm(earliestStart)

		// generate timestamps
		// every 5 mins for 7 days
		timestamps := make([]float64, 0)
		for offset := float64(0); offset < float64(60*60*24*co.days); offset += 5 * 60 {
			timestamps = append(timestamps, rootTs+offset)
		}

		timestampMap := map[int]int{}
		for i, ts := range timestamps {
			timestampMap[int(ts)] = i
		}

		// // do the dang thing
		cf := compFile{
			Name:      co.name,
			Marathons: marathonNames,
			Ts:        timestamps,
			Viewers:   make([][]*int64, 0),
			ViewersYt: make([][]*int64, 0),
			Donations: make([][]*float64, 0),
			Meta:      make([]meta, 0),
		}

		for _, marathon := range marathons {
			startTime := threePm(marathon.Games[0].Start)
			offsetSeconds := startTime - rootTs

			viewers := make([]*int64, len(timestamps))
			viewersYt := make([]*int64, len(timestamps))
			donations := make([]*float64, len(timestamps))
			meta := meta{}

			lastIndex := 0
			for _, v := range marathon.Viewers {
				// truncate to 5 mins
				truncatedTs := math.Floor(v.Time/(5*60)) * (5 * 60)
				offsetTs := int(truncatedTs - offsetSeconds)
				index, ok := timestampMap[offsetTs]
				if !ok {
					continue
				}

				setMax(&viewers[index], v.TwitchViewers)
				setMax(&viewersYt[index], v.YoutubeViewers)
				setMax(&donations[index], v.DonationTotal)

				if setMaxP(&meta.MaxViewers, v.TwitchViewers) {
					meta.MaxViewersTs = v.Time
				}
				setMaxP(&meta.MaxDonations, v.DonationTotal)

				lastIndex = index
			}

			// trim off the end
			viewers = viewers[:lastIndex]
			viewersYt = viewersYt[:lastIndex]

			// extend the donation count until the end
			for i := lastIndex; i < len(timestamps); i++ {
				donations[i] = donations[lastIndex]
			}

			// find the game at peak viewers
			for _, g := range marathon.Games {
				if g.Start > meta.MaxViewersTs {
					break
				}
				meta.MaxViewersGame = g.Name
			}

			cf.Viewers = append(cf.Viewers, viewers)
			cf.ViewersYt = append(cf.ViewersYt, viewersYt)
			cf.Donations = append(cf.Donations, donations)
			cf.Meta = append(cf.Meta, meta)
		}

		outputFile := args[0]
		b, err := json.MarshalIndent(cf, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(outputFile, b, 0o700); err != nil {
			return err
		}

		return nil
	},
}

func setMax[V cmp.Ordered](dst **V, v *V) {
	if v == nil {
		return
	}

	if *dst == nil || *v > **dst {
		x := *v
		*dst = &x
	}
}

func setMaxP[V cmp.Ordered](dst *V, v *V) bool {
	if v == nil {
		return false
	}

	if dst == nil || *v > *dst {
		*dst = *v
		return true
	}
	return false
}

func threePm(ts float64) float64 {
	t := tsToTime(ts)
	threepm := time.Date(t.Year(), t.Month(), t.Day(), 15, 0, 0, 0, time.UTC)
	return timeToTs(threepm)
}

func timeToTs(t time.Time) float64 {
	return float64(t.UnixMilli()) / 1000
}

func tsToTime(ts float64) time.Time {
	return time.UnixMilli(int64(ts * 1000))
}

func expand(globs []string) ([]string, error) {
	files := make([]string, 0)
	for _, g := range globs {
		matches, err := filepath.Glob(g)
		if err != nil {
			return files, err
		}
		files = append(files, matches...)
	}
	return files, nil
}

func sort(files []string) {
	orderIndex := []string{"agdq", "sgdq", "frost", "flame", "gdqx"}
	sortRe := regexp.MustCompile(`([a-z]+)(\d\d)`)
	slices.SortFunc(files, func(a string, b string) int {
		am := sortRe.FindStringSubmatch(filepath.Base(a))
		bm := sortRe.FindStringSubmatch(filepath.Base(b))
		if len(am) < 3 || len(bm) < 3 {
			return 0
		}

		akey := fmt.Sprintf("%s%d", am[2], slices.Index(orderIndex, am[1]))
		bkey := fmt.Sprintf("%s%d", bm[2], slices.Index(orderIndex, bm[1]))
		return cmp.Compare(bkey, akey)
	})
}

func init() {
	rootCmd.AddCommand(compCmd)

	compCmd.Flags().StringVar(&co.name, "name", "", "comparison name")
	compCmd.Flags().StringSliceVar(&co.globs, "glob", []string{}, "glob")
	compCmd.Flags().IntVar(&co.days, "days", 7, "number of days")

	compCmd.MarkFlagRequired("name")
}
