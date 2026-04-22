package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const userAgent = "Mozilla/5.0 (compatible; DLSKitScraper/1.0; +https://example.com/bot)"

type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 8 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
	}
}

func (h *HTTPClient) Get(ctx context.Context, rawURL string) (string, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := h.client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", resp.StatusCode, fmt.Errorf("GET %s returned %d", rawURL, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 12<<20))
	if err != nil {
		return "", resp.StatusCode, err
	}

	return string(body), resp.StatusCode, nil
}

func (h *HTTPClient) ValidPNG(ctx context.Context, rawURL string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "image/png,*/*;q=0.8")

	resp, err := h.client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return true
		}
		if resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusForbidden {
			return false
		}
	}

	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return false
	}
	getReq.Header.Set("User-Agent", userAgent)
	getReq.Header.Set("Accept", "image/png,*/*;q=0.8")

	getResp, err := h.client.Do(getReq)
	if err != nil {
		return false
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		return false
	}

	contentType := strings.ToLower(getResp.Header.Get("Content-Type"))
	if contentType == "" {
		return true
	}
	return strings.Contains(contentType, "image/png") || strings.Contains(contentType, "application/octet-stream")
}
