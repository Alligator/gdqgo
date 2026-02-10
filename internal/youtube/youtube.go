package youtube

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/alligator/gdqgo/internal/persist"
)

type liveStreamingDetails struct {
	ConcurrentViewers *int `json:"concurrentViewers,string"`
}

type videoResponseItem struct {
	LiveStreamingDetails liveStreamingDetails `json:"liveStreamingDetails"`
}

type videoResponse struct {
	Items []videoResponseItem `json:"items"`
}

type searchResponse struct {
	Items []searchResponseItem `json:"items"`
}

type searchResponseItem struct {
	Id struct {
		VideoId string `json:"videoId"`
	} `json:"id"`
}

type ViewerResult struct {
	Viewers int
	Live    bool
}

func GetLiveVideoId(channelId string, apiKey string) (string, error) {
	qp := url.Values{}
	qp.Add("part", "id")
	qp.Add("channelId", channelId)
	qp.Add("eventType", "live")
	qp.Add("type", "video")
	qp.Add("maxResults", "1")
	qp.Add("key", apiKey)
	qp.Add("fields", "items/id/videoId")

	resp, err := http.Get("https://www.googleapis.com/youtube/v3/search?" + qp.Encode())
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("youtube /search returned HTTP %s", resp.Status)
	}

	var r searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}

	if len(r.Items) == 0 {
		return "", nil
	}

	return r.Items[0].Id.VideoId, nil
}

func GetViewers(channelId string) (ViewerResult, error) {
	res := ViewerResult{}

	apiKey, err := persist.GetExpected("youtube_api_key")
	if err != nil {
		return res, err
	}

	videoId, ok, err := persist.Get("youtube_video_id")
	if err != nil {
		return res, err
	}
	if !ok || len(videoId) == 0 {
		videoId, err = GetLiveVideoId(channelId, apiKey)
		if err != nil {
			return res, err
		}
		persist.Set("youtube_video_id", videoId)
		if len(videoId) == 0 {
			// channel is not live
			return res, nil
		}
	}

	qp := url.Values{}
	qp.Add("id", videoId)
	qp.Add("key", apiKey)
	qp.Add("part", "liveStreamingDetails")

	resp, err := http.Get("https://www.googleapis.com/youtube/v3/videos?" + qp.Encode())
	if err != nil {
		return res, err
	}
	if resp.StatusCode != http.StatusOK {
		return res, fmt.Errorf("youtube /videos returned HTTP %s", resp.Status)
	}

	var r videoResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return res, err
	}

	if len(r.Items) == 0 {
		return res, fmt.Errorf("no videos found")
	}

	if r.Items[0].LiveStreamingDetails.ConcurrentViewers == nil {
		// stream is probably not live, clear the cached id
		persist.Set("youtube_video_id", "")
		return res, nil
	}

	res.Viewers = *r.Items[0].LiveStreamingDetails.ConcurrentViewers
	res.Live = true
	return res, nil
}
