package kaggle

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"
)

const BaseURL = "https://www.kaggle.com/api/v1"

type Client struct {
	Username string
	Key      string
	HTTP     *http.Client
}

func NewClient(username, key string) *Client {
	return &Client{
		Username: username,
		Key:      key,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, endpoint string, body io.Reader, contentType string) ([]byte, error) {
	req, err := http.NewRequest(method, BaseURL+endpoint, body)
	if err != nil {
		return nil, err
	}

	// Basic Auth
	auth := base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.Key))
	req.Header.Add("Authorization", "Basic "+auth)
	
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("kaggle api error: %s - %s", resp.Status, string(respBody))
	}

	return respBody, nil
}
