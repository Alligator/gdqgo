package youtube

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/alligator/gdqgo/internal/persist"
)

type liveStreamingDetails struct {
	ConcurrentViewers int `json:"concurrentViewers,string"`
}

type videoResponseItem struct {
	LiveStreamingDetails liveStreamingDetails `json:"liveStreamingDetails"`
}

type videoResponse struct {
	Items []videoResponseItem `json:"items"`
}

func GetViewers(videoId string) (int, error) {
	apiKey, err := persist.GetExpected("youtube_api_key")
	if err != nil {
		return 0, err
	}

	qp := url.Values{}
	qp.Add("id", videoId)
	qp.Add("key", apiKey)
	qp.Add("part", "liveStreamingDetails")

	resp, err := http.Get("https://www.googleapis.com/youtube/v3/videos?" + qp.Encode())
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("youtube returned HTTP %s", resp.Status)
	}

	var r videoResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return 0, err
	}

	if len(r.Items) == 0 {
		return 0, fmt.Errorf("no videos found")
	}

	return r.Items[0].LiveStreamingDetails.ConcurrentViewers, nil
}
