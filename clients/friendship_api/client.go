package friendship_api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxBodySize = 1 << 20

type FriendshipClient struct {
	baseURL string
	client  *http.Client
}

func NewFriendshipClient(baseURL string, httpClient *http.Client) (*FriendshipClient, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	return &FriendshipClient{baseURL: baseURL, client: httpClient}, nil
}

func (c *FriendshipClient) GetRelationship(ctx context.Context, viewerID, targetID uuid.UUID) (RelationshipResponse, error) {
	endpoint, err := url.JoinPath(c.baseURL, "friends", targetID.String(), "relationship")

	if err != nil {

		return RelationshipResponse{}, fmt.Errorf("build endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)

	if err != nil {

		return RelationshipResponse{}, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("X-User-ID", viewerID.String())

	resp, err := c.client.Do(req)

	if err != nil {

		return RelationshipResponse{}, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))

		return RelationshipResponse{}, fmt.Errorf("friendship service error (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var out RelationshipResponse

	if err = json.NewDecoder(io.LimitReader(resp.Body, maxBodySize)).Decode(&out); err != nil {

		return RelationshipResponse{}, fmt.Errorf("decode json: %w", err)
	}

	return out, nil

}
