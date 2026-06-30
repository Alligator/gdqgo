package fetch

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/alligator/gdqgo/internal/logger"
	"github.com/alligator/gdqgo/internal/statsfile"
	"github.com/alligator/gdqgo/internal/tracker"
	"github.com/alligator/gdqgo/internal/twitch"
	"github.com/alligator/gdqgo/internal/youtube"
)

type FetchOpts struct {
	Name               string
	Typ                string
	TrackerMararthonId int
	TwitchUserId       int
	YoutubeChannelId   string
	Step               string
}

type Fetcher struct {
	errs []error
	opts FetchOpts
}

func NewFetcher(fo FetchOpts) Fetcher {
	return Fetcher{make([]error, 0), fo}
}

func (f *Fetcher) step(name string, fn func() error) {
	defer func() {
		if r := recover(); r != nil {
			f.errs = append(f.errs, fmt.Errorf("[%s] PANIC: %v\n%s", name, r, debug.Stack()))
		}
	}()

	if f.opts.Step != "" && name != f.opts.Step {
		return
	}

	start := time.Now()
	logger.Debugf("fetch", "[%s] starting", name)

	if err := fn(); err != nil {
		f.errs = append(f.errs, fmt.Errorf("[%s] ERROR: %w", name, err))
		logger.Logf("fetch", "[%s] ERROR: %s", name, err)
	}

	d := time.Since(start).Round(time.Millisecond)
	logger.Debugf("fetch", "[%s] took %s", name, d)
}

func (f *Fetcher) DoFetch(file string) error {
	var sf statsfile.StatsFile
	sf, err := statsfile.Read(file)
	if err != nil {
		if os.IsNotExist(err) {
			sf = statsfile.New(f.opts.Name, f.opts.Typ)
		} else {
			return err
		}
	}

	f.step("fetch schedule", func() error {
		schedule, err := tracker.GetSchedule(f.opts.TrackerMararthonId)
		if err != nil {
			return err
		}
		logger.Debugf("fetch", "fetched schedule")
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

		logger.Debugf("fetch", "fetched %d games", len(games))
		sf.Games = games
		return nil
	})

	v := statsfile.Viewer{}
	now := time.Now().UTC()
	v.Time = float64(now.UnixMilli()) / 1000

	f.step("fetch donations", func() error {
		donations, err := tracker.GetDonations(f.opts.TrackerMararthonId)
		if err != nil {
			return err
		}
		logger.Debugf("fetch", "fetched donation total")
		v.DonationTotal = &donations
		return nil
	})

	f.step("fetch twitch viewers", func() error {
		viewers, err := twitch.GetViewers(f.opts.TwitchUserId)
		if err != nil {
			return err
		}
		logger.Debugf("fetch", "fetched twitch viewers")
		i64 := int64(viewers)
		v.TwitchViewers = &i64
		return nil
	})

	f.step("fetch youtube viewers", func() error {
		viewers, err := youtube.GetViewers(f.opts.YoutubeChannelId)
		if err != nil {
			return err
		}
		logger.Debugf("fetch", "fetched youtube viewers")
		i64 := int64(0)
		if viewers.Live {
			i64 = int64(viewers.Viewers)
		}
		v.YoutubeViewers = &i64
		return nil
	})

	sf.Viewers = append(sf.Viewers, v)

	if err := statsfile.Write(file, sf); err != nil {
		f.errs = append(f.errs, err)
	}

	if len(f.errs) > 0 {
		return errors.Join(f.errs...)
	}

	return nil
}
