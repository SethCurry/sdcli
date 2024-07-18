package stability

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// ClientOption is a function that can be used as an option
// for the NewClient function.
type ClientOption func(*Client)

// NewClient creates a new *Client.  The API key is mandatory,
// and any number of ClientOptions can be passed in to customize
// the client's behavior.
func NewClient(apiKey string, options ...ClientOption) *Client {
	client := &Client{
		apiKey:     apiKey,
		baseURL:    "https://api.stability.ai",
		httpClient: http.DefaultClient,
	}

	for _, option := range options {
		option(client)
	}

	return client
}

// Client implements a client for Stability's Stable Diffusion API.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func (c *Client) Generate3(ctx context.Context, writeTo io.Writer, generateRequest Generate3Request) error {
	if err := generateRequest.Validate(); err != nil {
		return fmt.Errorf("Generate3Request is invalid: %w", err)
	}

	reqURL := fmt.Sprintf("%s/v2beta/stable-image/generate/sd3", c.baseURL)

	var formBuf bytes.Buffer

	contentType, err := generateRequest.ToFormData(&formBuf)
	if err != nil {
		return fmt.Errorf("failed to generate form data for Generate3 request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, &formBuf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "image/*")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("got unexpected status code %d while generating image. Response: %s", resp.StatusCode, string(body))
	}

	_, err = io.Copy(writeTo, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy image to writer: %w", err)
	}

	return nil
}
