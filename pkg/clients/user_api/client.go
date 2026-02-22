package user_api

import (
	"bytes"
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

type UserClient struct {
	baseURL string
	client  *http.Client
}

func NewUserClient(baseURL string, httpClient *http.Client) (*UserClient, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}

	if httpClient == nil {
		httpClient = &http.Client{ //fixme parse from config
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       100,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: false,
			},
		}
	}

	return &UserClient{
		baseURL: baseURL,
		client:  httpClient,
	}, nil

}

func (c *UserClient) CreateUser(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	endpoint, err := url.JoinPath(c.baseURL, "user")

	if err != nil {

		return nil, fmt.Errorf("build users endpoint: %w", err)
	}

	body, err := json.Marshal(req)

	if err != nil {

		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))

	if err != nil {

		return nil, fmt.Errorf("build request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)

	if err != nil {

		return nil, fmt.Errorf("execute request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {

		return nil, decodeError(resp)
	}

	return decodeUser(resp)
}

func (c *UserClient) AuthenticateUser(ctx context.Context, email, password string) (*UserResponse, error) {
	endpoint, err := url.JoinPath(c.baseURL, "users", "authenticate-user")

	if err != nil {

		return nil, fmt.Errorf("build users endpoint: %w", err)
	}

	body := map[string]string{
		"email":        email,
		"passwordHash": password,
	}

	bodyBytes, err := json.Marshal(body)

	if err != nil {

		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))

	if err != nil {

		return nil, fmt.Errorf("build request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)

	if err != nil {

		return nil, fmt.Errorf("execute request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {

		return nil, decodeError(resp)
	}

	return decodeUser(resp)
}

func (c *UserClient) GetUserByID(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	endpoint, err := url.JoinPath(c.baseURL, "users", id.String())

	if err != nil {

		return nil, fmt.Errorf("build users endpoint: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)

	if err != nil {

		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.client.Do(httpReq)

	if err != nil {

		return nil, fmt.Errorf("execute request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, decodeError(resp)
	}

	return decodeUser(resp)
}

func (c *UserClient) FindByEmail(ctx context.Context, email string) (*UserResponse, error) {
	endpoint, err := url.JoinPath(c.baseURL, "users", "by-email")
	if err != nil {
		return nil, fmt.Errorf("build users endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-User-Email", email)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, decodeError(resp)
	}

	return decodeUser(resp)
}

func decodeUser(resp *http.Response) (*UserResponse, error) {

	var out UserResponse

	dec := json.NewDecoder(io.LimitReader(resp.Body, maxBodySize))
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	return &out, nil
}

func decodeError(resp *http.Response) error {
	data, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	msg := strings.TrimSpace(string(data))

	switch resp.StatusCode {
	case http.StatusBadRequest:
		if msg == "" {
			return fmt.Errorf("bad request")
		}
		return fmt.Errorf("bad request: %s", msg)
	case http.StatusNotFound:
		return ErrNotFound
	default:
		if msg == "" {
			return fmt.Errorf("server error (%d)", resp.StatusCode)
		}
		return fmt.Errorf("server error (%d): %s", resp.StatusCode, msg)
	}
}
