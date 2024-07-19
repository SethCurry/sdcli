// Package stability implements a client for the Stable Diffusion API.
package stability

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
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

type requestFormData interface {
	toFormData(*multipart.Writer) error
}

func (c *Client) newRequest(ctx context.Context, reqURL string, requestData requestFormData) (*http.Request, error) {
	var formData bytes.Buffer

	writer := multipart.NewWriter(&formData)

	if err := requestData.toFormData(writer); err != nil {
		writer.Close()
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close form data writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, &formData)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "image/*")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

func (c *Client) doRequest(req *http.Request, writeTo io.Writer) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("expected status code 200 on HTTP request, but got %d", resp.StatusCode)
	}

	if _, err := io.Copy(writeTo, resp.Body); err != nil {
		return fmt.Errorf("failed to copy image response: %w", err)
	}

	return nil
}

// GenerateUltra generates an image using the Stable Image Ultra API.
func (c *Client) GenerateUltra(ctx context.Context, writeTo io.Writer, generateRequest GenerateUltraRequest) error {
	if err := generateRequest.validate(); err != nil {
		return err
	}

	reqURL := fmt.Sprintf("%s/v2beta/stable-image/generate/ultra", c.baseURL)

	req, err := c.newRequest(ctx, reqURL, generateRequest)
	if err != nil {
		return err
	}

	return c.doRequest(req, writeTo)
}

// Generate3 generates an image using the Stable Diffusion 3 API.
//
// API Reference: https://platform.stability.ai/docs/api-reference#tag/Generate/paths/~1v2beta~1stable-image~1generate~1sd3/post
func (c *Client) Generate3(ctx context.Context, writeTo io.Writer, generateRequest Generate3Request) error {
	if err := generateRequest.validate(); err != nil {
		return fmt.Errorf("Generate3Request is invalid: %w", err)
	}

	reqURL := fmt.Sprintf("%s/v2beta/stable-image/generate/sd3", c.baseURL)

	req, err := c.newRequest(ctx, reqURL, generateRequest)
	if err != nil {
		return err
	}

	return c.doRequest(req, writeTo)
}
