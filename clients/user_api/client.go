package user_api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/common_service/apperror"
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
	endpoint, err := url.JoinPath(c.baseURL, "users")

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

func (c *UserClient) GetUserByID(ctx context.Context, id uuid.UUID) (*UserResponse, error) { //fixme
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

func (c *UserClient) GetUserByEmail(ctx context.Context, email string) (*UserResponse, error) {
	endpoint, err := url.JoinPath(c.baseURL, "users", "by-email")

	if err != nil {
		return nil, fmt.Errorf("build users endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)

	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	q := req.URL.Query()
	q.Set("email", email)
	req.URL.RawQuery = q.Encode()
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

func (c *UserClient) GetUserByUsername(ctx context.Context, username string) (*UserResponse, error) {
	endpoint, err := url.JoinPath(c.baseURL, "users", "by-username", username)

	if err != nil {
		return nil, fmt.Errorf("build users endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)

	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

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
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))

	var errResp struct {
		Code  string `json:"code"`
		Field string `json:"field"`
	}
	_ = json.Unmarshal(body, &errResp)

	switch resp.StatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusConflict:
		if errResp.Code != "" {
			return apperror.Conflict(errResp.Code, errResp.Field, "conflict")
		}
		return apperror.Conflict(apperror.CodeAlreadyExists, "", "resource already exists")
	case http.StatusBadRequest:
		if errResp.Code != "" {
			return apperror.BadRequest(errResp.Code, errResp.Field)
		}
		return apperror.BadRequest(apperror.CodeValidationFailed, strings.TrimSpace(string(body)))
	case http.StatusUnprocessableEntity:
		if errResp.Code != "" {
			return apperror.Validation(errResp.Code, errResp.Field, "validation failed")
		}
		return apperror.BadRequest(apperror.CodeValidationFailed, "validation failed")
	default:
		return apperror.Internal(
			fmt.Sprintf("upstream error (%d)", resp.StatusCode),
			fmt.Errorf("%s", strings.TrimSpace(string(body))),
		)
	}
}

func (c *UserClient) ActivateUser(ctx context.Context, email string) error {
	endpoint, err := url.JoinPath(c.baseURL, "users", "activate")

	if err != nil {
		return fmt.Errorf("build activate endpoint: %w", err)
	}

	body, err := json.Marshal(map[string]string{"email": email})

	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))

	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)

	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeError(resp)
	}
	return nil
}
