package cbr

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	url        string
	fallback   string
}

func NewClient(url string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		url:      url,
		fallback: "16.0000",
	}
}

func (c *Client) GetKeyRate(ctx context.Context) (string, error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.url,
		bytes.NewBufferString(keyRateSOAPEnvelope()),
	)
	if err != nil {
		return "", fmt.Errorf("create cbr request: %w", err)
	}

	request.Header.Set("Content-Type", "text/xml; charset=utf-8")
	request.Header.Set("SOAPAction", "http://web.cbr.ru/KeyRate")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return c.fallback, nil
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return c.fallback, nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("read cbr response: %w", err)
	}

	rate, err := extractKeyRate(string(body))
	if err != nil {
		return c.fallback, nil
	}

	return rate, nil
}
