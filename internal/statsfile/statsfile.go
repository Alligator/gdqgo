package statsfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Game struct {
	Start    float64
	Name     string
	Runners  string
	Category string
}

type Viewer struct {
	Time           float64
	TwitchViewers  *int64
	YoutubeViewers *int64
	DonationTotal  *float64
}

type StatsFile struct {
	MarathonName string   `json:"marathon_name"`
	MarathonType string   `json:"marathon_type"`
	Viewers      []Viewer `json:"viewers"`
	Games        []Game   `json:"games"`
	Filename     string
}

func New(name string, typ string) StatsFile {
	return StatsFile{
		MarathonName: name,
		MarathonType: typ,
		Viewers:      make([]Viewer, 0),
		Games:        make([]Game, 0),
	}
}

func Read(path string) (StatsFile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return StatsFile{}, err
	}

	var sf StatsFile
	if err := json.Unmarshal(b, &sf); err != nil {
		return StatsFile{}, err
	}

	sf.Filename = filepath.Base(path)
	return sf, nil
}

func Write(path string, sf StatsFile) error {
	b, err := json.Marshal(sf)
	if err != nil {
		return err
	}

	tmpPath := path + ".tnp"
	if err := os.WriteFile(tmpPath, b, 0o664); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	return nil
}

func (v Viewer) MarshalJSON() ([]byte, error) {
	row := []any{
		v.Time,
		valueOrNil(v.TwitchViewers),
		valueOrNil(v.DonationTotal),
		valueOrNil(v.YoutubeViewers),
	}
	return json.Marshal(row)
}

func valueOrNil[V any](v *V) any {
	if v == nil {
		return nil
	}
	return v
}

func (v *Viewer) UnmarshalJSON(b []byte) error {
	var row []json.RawMessage
	if err := json.Unmarshal(b, &row); err != nil {
		return err
	}

	if err := json.Unmarshal(row[0], &v.Time); err != nil {
		return fmt.Errorf("viewer[0] timestamp %w", err)
	}
	if err := json.Unmarshal(row[1], &v.TwitchViewers); err != nil {
		return fmt.Errorf("viewer[1] twitch viewers %w", err)
	}
	if err := json.Unmarshal(row[2], &v.DonationTotal); err != nil {
		return fmt.Errorf("viewer[2] donation total %w", err)
	}

	if len(row) > 3 {
		if err := json.Unmarshal(row[3], &v.YoutubeViewers); err != nil {
			return fmt.Errorf("viewer[3] youtube viewers %w", err)
		}
	}

	return nil
}

func (g Game) MarshalJSON() ([]byte, error) {
	row := []any{g.Start, g.Name, g.Runners, g.Category}
	return json.Marshal(row)
}

func (g *Game) UnmarshalJSON(b []byte) error {
	var row []json.RawMessage
	if err := json.Unmarshal(b, &row); err != nil {
		return err
	}

	if err := json.Unmarshal(row[0], &g.Start); err != nil {
		return fmt.Errorf("game[0] start %w", err)
	}
	if err := json.Unmarshal(row[1], &g.Name); err != nil {
		return fmt.Errorf("game[1] name %w", err)
	}

	if len(row) > 2 {
		if err := json.Unmarshal(row[2], &g.Runners); err != nil {
			return fmt.Errorf("game[2] runners %w", err)
		}
	}

	if len(row) > 3 {
		if err := json.Unmarshal(row[3], &g.Category); err != nil {
			return fmt.Errorf("game[3] category %w", err)
		}
	}

	return nil
}
