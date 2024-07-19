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

// GenerateUltra generates an image using the Stable Image Ultra API.
func (c *Client) GenerateUltra(ctx context.Context, writeTo io.Writer, generateRequest GenerateUltraRequest) error {
	if err := generateRequest.validate(); err != nil {
		return err
	}

	reqURL := fmt.Sprintf("%s/v2beta/stable-image/generate/ultra", c.baseURL)

	var formBuf bytes.Buffer

	formWriter := multipart.NewWriter(&formBuf)

	if err := generateRequest.toFormData(formWriter); err != nil {
		formWriter.Close()
		return fmt.Errorf("failed to generate form data for request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, &formBuf)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", formWriter.FormDataContentType())
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

// Generate3 generates an image using the Stable Diffusion 3 API.
//
// API Reference: https://platform.stability.ai/docs/api-reference#tag/Generate/paths/~1v2beta~1stable-image~1generate~1sd3/post
func (c *Client) Generate3(ctx context.Context, writeTo io.Writer, generateRequest Generate3Request) error {
	if err := generateRequest.validate(); err != nil {
		return fmt.Errorf("Generate3Request is invalid: %w", err)
	}

	reqURL := fmt.Sprintf("%s/v2beta/stable-image/generate/sd3", c.baseURL)

	var formBuf bytes.Buffer

	formWriter := multipart.NewWriter(&formBuf)

	err := generateRequest.toFormData(formWriter)
	if err != nil {
		formWriter.Close()
		return fmt.Errorf("failed to generate form data for Generate3 request: %w", err)
	}

	if err = formWriter.Close(); err != nil {
		return fmt.Errorf("failed to close form writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, &formBuf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", formWriter.FormDataContentType())
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
