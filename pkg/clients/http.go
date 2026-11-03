package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTP is a small JSON-oriented client with a mandatory timeout.
type HTTP struct {
	client *http.Client
}

// NewHTTP returns a client with the given timeout (defaults to 10s).
func NewHTTP(timeout time.Duration) *HTTP {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &HTTP{
		client: &http.Client{Timeout: timeout},
	}
}

// PostJSON posts a JSON body and returns status + raw body.
func (c *HTTP) PostJSON(ctx context.Context, url string, payload any) (int, []byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, data, nil
}

// PostJSONExpect posts JSON and errors unless status matches want.
func (c *HTTP) PostJSONExpect(ctx context.Context, url string, payload any, want int) ([]byte, error) {
	status, data, err := c.PostJSON(ctx, url, payload)
	if err != nil {
		return nil, err
	}
	if status != want {
		return data, fmt.Errorf("unexpected status %d from %s: %s", status, url, string(data))
	}
	return data, nil
}
