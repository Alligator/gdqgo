package tracker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Runner struct {
	Name string
}

type Run struct {
	Name      string `json:"display_name"`
	RunType   string `json:"type"`
	Category  string
	Runners   []Runner
	StartTime time.Time `json:"starttime"`
}

type Runs struct {
	Results []Run
}

func GetSchedule(trackerMarathonId int) ([]Run, error) {
	url := fmt.Sprintf("https://tracker.gamesdonequick.com/tracker/api/v2/events/%d/runs", trackerMarathonId)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	var result Runs
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Results, nil
}

type event struct {
	DonationTotal float64 `json:"donation_total"`
}

func GetDonations(trackerMarathonId int) (float64, error) {
	url := fmt.Sprintf("https://tracker.gamesdonequick.com/tracker/api/v2/events/%d/?totals", trackerMarathonId)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	var result event
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0, err
	}

	return result.DonationTotal, nil
}
