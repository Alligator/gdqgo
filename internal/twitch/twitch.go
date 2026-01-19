package twitch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/alligator/gdqgo/internal/persist"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpresIn    int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type stream struct {
	ViewerCount int `json:"viewer_count"`
}

type streams struct {
	Data []stream
}

func getTwitchAccessToken() (string, error) {
	clientId, err := persist.GetExpected("twitch_client_id")
	if err != nil {
		return "", err
	}

	clientSecret, err := persist.GetExpected("twitch_client_secret")
	if err != nil {
		return "", err
	}

	u, _ := url.Parse("https://id.twitch.tv/oauth2/token")

	qp := url.Values{}
	qp.Add("client_id", clientId)
	qp.Add("client_secret", clientSecret)
	qp.Add("grant_type", "client_credentials")

	u.RawQuery = qp.Encode()

	resp, err := http.Post(u.String(), "text/plain", nil)
	if err != nil {
		return "", nil
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("twitch returned HTTP %s", resp.Status)
	}

	var token tokenResponse
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func GetViewers(userId int) (int, error) {
	twitchToken, ok, err := persist.Get("twitch_token")
	if err != nil {
		return 0, err
	}

	clientId, err := persist.GetExpected("twitch_client_id")
	if err != nil {
		return 0, err
	}

	if !ok {
		twitchToken, err = getTwitchAccessToken()
		if err != nil {
			return 0, err
		}
		persist.Set("twitch_token", twitchToken)
	}

	client := http.Client{}

	url := fmt.Sprintf("https://api.twitch.tv/helix/streams?user_id=%d", userId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("Client-ID", clientId)
	req.Header.Add("Authorization", "Bearer "+twitchToken)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("twitch returned HTTP %s", resp.Status)
	}

	var s streams
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return 0, err
	}

	if len(s.Data) == 0 {
		return 0, fmt.Errorf("twitch returned no data")
	}

	return s.Data[0].ViewerCount, nil
}
