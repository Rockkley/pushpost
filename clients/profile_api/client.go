package profile_api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxBodySize = 1 << 20

type ProfileClient struct {
	baseURL string
	client  *http.Client
}

func NewProfileClient(baseURL string, httpClient *http.Client) (*ProfileClient, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &ProfileClient{baseURL: baseURL, client: httpClient}, nil
}

func (c *ProfileClient) GetByUsername(ctx context.Context, username string) (ProfileResponse, error) {
	endpoint, err := url.JoinPath(c.baseURL, "profiles", "by-username", username)
	if err != nil {
		return ProfileResponse{}, fmt.Errorf("build endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ProfileResponse{}, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return ProfileResponse{}, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
		if resp.StatusCode == http.StatusNotFound {
			return ProfileResponse{}, ErrNotFound
		}
		return ProfileResponse{}, fmt.Errorf("profile service error (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var out ProfileResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBodySize)).Decode(&out); err != nil {
		return ProfileResponse{}, fmt.Errorf("decode json: %w", err)
	}
	return out, nil
}
