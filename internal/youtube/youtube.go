package youtube

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/alligator/gdqgo/internal/logger"
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

func canFetchVideoId() (bool, error) {
	lastFetchStr, ok, err := persist.Get("youtube_last_fetch")
	if err != nil {
		return false, err
	}

	if ok {
		t, err := strconv.ParseInt(lastFetchStr, 10, 64)
		if err != nil {
			// clear the value and continue the fetch
			persist.Set("youtube_last_fetch", "")
			return true, nil
		}

		lastVideoIdFetch := time.Unix(t, 0)
		since := time.Since(lastVideoIdFetch)

		if since < time.Minute*15 {
			logger.Debugf("youtube", "skipped fetching a new id, only %s elapsed", since.String())
			return false, nil
		}
	}

	return true, nil
}

func GetLiveVideoId(channelId string, apiKey string) (string, error) {
	canFetch, err := canFetchVideoId()
	if err != nil {
		return "", err
	}

	if !canFetch {
		return "", nil
	}

	logger.Debugf("youtube", "fetching new video id")

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
	logger.Debugf("youtube", "fetched video id")

	now := time.Now()
	persist.Set("youtube_last_fetch", strconv.Itoa(int(now.Unix())))

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

	logger.Debugf("youtube", "fetching viewers for video %s", videoId)
	resp, err := http.Get("https://www.googleapis.com/youtube/v3/videos?" + qp.Encode())
	if err != nil {
		return res, err
	}
	if resp.StatusCode != http.StatusOK {
		return res, fmt.Errorf("youtube /videos returned HTTP %s", resp.Status)
	}
	logger.Debugf("youtube", "fetched viewers for video %s", videoId)

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
