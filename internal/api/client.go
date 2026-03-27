package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coinmarketcap/coinmarketcap-cli/internal/config"
)

const (
	maxErrorBodySize    = 1 << 20  // 1MB
	maxResponseBodySize = 50 << 20 // 50MB — guards against pathological upstream responses
)

var (
	ErrInvalidAPIKey     = fmt.Errorf("invalid API key — check your key with `cmc status` or set a new one with `cmc auth`")
	ErrPlanRestricted    = fmt.Errorf("this endpoint requires a higher-tier plan — visit https://coinmarketcap.com/api/pricing/")
	ErrRateLimited       = fmt.Errorf("rate limited — please wait and try again")
	ErrAssetNotFound     = fmt.Errorf("asset not found")
	ErrResolverAmbiguous = fmt.Errorf("multiple assets match symbol")
	ErrInvalidInput      = fmt.Errorf("invalid request parameters")
)

// RateLimitError carries rate-limit metadata from a 429 response.
type RateLimitError struct {
	RetryAfter int       // seconds from Retry-After header; 0 if absent
	ResetAt    time.Time // from x-ratelimit-reset header; zero if absent
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limited — retry after %d seconds", e.RetryAfter)
	}
	if !e.ResetAt.IsZero() {
		return fmt.Sprintf("rate limited — resets at %s", e.ResetAt.UTC().Format("15:04:05 UTC"))
	}
	return "rate limited — please wait and try again"
}

func (e *RateLimitError) Is(target error) bool {
	return target == ErrRateLimited
}

// apiErrorResponse covers CMC's error JSON format:
//
//	{"status": {"error_code": N, "error_message": "..."}}
type apiErrorResponse struct {
	Status *struct {
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	} `json:"status"`
}

// extractMessage returns the error message from the response body.
func (e *apiErrorResponse) extractMessage() string {
	if e.Status != nil && e.Status.ErrorMessage != "" {
		return e.Status.ErrorMessage
	}
	return ""
}

// classify401 determines whether a 401 response is an auth failure.
// CMC returns 401 for invalid/missing API keys (error codes 1001, 1002).
func classify401(apiErr apiErrorResponse, msg string) error {
	code := 0
	if apiErr.Status != nil {
		code = apiErr.Status.ErrorCode
	}
	lower := strings.ToLower(msg)

	// CMC auth failures: error codes 1001 (invalid), 1002 (missing)
	if code == 1001 || code == 1002 || isAuthFailure(lower) {
		if msg != "" {
			return fmt.Errorf("%s (%w)", msg, ErrInvalidAPIKey)
		}
		return ErrInvalidAPIKey
	}

	// Unknown 401 — return generic error or assume auth
	if msg != "" {
		return fmt.Errorf("API error 401: %s", msg)
	}
	return ErrInvalidAPIKey
}

func isAuthFailure(lowerMsg string) bool {
	for _, phrase := range []string{
		"invalid api key",
		"api key missing",
		"api key is invalid",
	} {
		if strings.Contains(lowerMsg, phrase) {
			return true
		}
	}
	return false
}

func isAssetNotFound(lowerMsg string) bool {
	for _, phrase := range []string{
		"no data found",
		"not found",
		"could not find",
	} {
		if strings.Contains(lowerMsg, phrase) {
			return true
		}
	}
	return false
}

func isInvalidInput(lowerMsg string) bool {
	for _, phrase := range []string{
		"invalid",
		"required",
	} {
		if strings.Contains(lowerMsg, phrase) {
			return true
		}
	}
	return false
}

type Client struct {
	http       *http.Client
	baseURLVal string // override; empty = use cfg.BaseURL()
	cfg        *config.Config
	UserAgent  string // sent with every request; set by cmd layer
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		http: &http.Client{Timeout: 30 * time.Second},
		cfg:  cfg,
	}
}

func NewClientWithHTTP(cfg *config.Config, httpClient *http.Client) *Client {
	return &Client{
		http: httpClient,
		cfg:  cfg,
	}
}

func (c *Client) SetBaseURL(url string) {
	c.baseURLVal = url
}

func (c *Client) baseURL() string {
	if c.baseURLVal != "" {
		return c.baseURLVal
	}
	return c.cfg.BaseURL()
}

func (c *Client) get(ctx context.Context, path string, result any) error {
	url := c.baseURL() + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	c.cfg.ApplyAuth(req)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return c.handleError(resp)
	}

	lr := io.LimitReader(resp.Body, maxResponseBodySize+1)
	dec := json.NewDecoder(lr)
	if err := dec.Decode(result); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return nil
}

func (c *Client) handleError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))

	// Parse CMC error body.
	var apiErr apiErrorResponse
	_ = json.Unmarshal(body, &apiErr)
	msg := apiErr.extractMessage()
	lower := strings.ToLower(msg)

	switch resp.StatusCode {
	case http.StatusBadRequest:
		switch {
		case isAssetNotFound(lower):
			if msg != "" {
				return fmt.Errorf("%s (%w)", msg, ErrAssetNotFound)
			}
			return ErrAssetNotFound
		case isInvalidInput(lower):
			if msg != "" {
				return fmt.Errorf("%s (%w)", msg, ErrInvalidInput)
			}
			return ErrInvalidInput
		default:
			if msg != "" {
				return fmt.Errorf("API error 400: %s", msg)
			}
			return fmt.Errorf("API error 400: %s", string(body))
		}

	case http.StatusUnauthorized:
		return classify401(apiErr, msg)

	case http.StatusForbidden:
		// CMC returns 403 for plan restrictions (error code 1003)
		if msg != "" {
			return fmt.Errorf("%s (%w)", msg, ErrPlanRestricted)
		}
		return ErrPlanRestricted

	case http.StatusNotFound:
		if isAssetNotFound(lower) {
			if msg != "" {
				return fmt.Errorf("%s (%w)", msg, ErrAssetNotFound)
			}
			return ErrAssetNotFound
		}
		if msg != "" {
			return fmt.Errorf("API error 404: %s", msg)
		}
		return fmt.Errorf("API error 404: %s", string(body))

	case http.StatusTooManyRequests:
		rle := &RateLimitError{}
		if retry := resp.Header.Get("Retry-After"); retry != "" {
			if secs, err := strconv.Atoi(retry); err == nil && secs > 0 {
				rle.RetryAfter = secs
			}
		}
		if reset := resp.Header.Get("x-ratelimit-reset"); reset != "" {
			// CMC may send timestamp format
			if t, err := time.Parse("2006-01-02 15:04:05 -0700", reset); err == nil {
				rle.ResetAt = t
			}
		}
		return rle

	default:
		if msg != "" {
			return fmt.Errorf("API error %d: %s", resp.StatusCode, msg)
		}
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
}

func (c *Client) requirePaid() error {
	if !c.cfg.IsPaid() {
		return ErrPlanRestricted
	}
	return nil
}
