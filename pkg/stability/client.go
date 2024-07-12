package stability

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

type Generate3Request struct {
	AspectRatio    string
	Prompt         string
	Model          string
	OutputFormat   string
	NegativePrompt string
	Strength       float32
	Image          io.Reader
}

func (g Generate3Request) ToFormData(writeTo io.Writer) (string, error) {
	writer := multipart.NewWriter(writeTo)

	if err := writer.WriteField("aspect_ratio", g.AspectRatio); err != nil {
		return "", fmt.Errorf("failed to write aspect_ratio field in form data: %w", err)
	}

	if err := writer.WriteField("prompt", g.Prompt); err != nil {
		return "", fmt.Errorf("failed to write prompt field in form data: %w", err)
	}

	if g.Model != "" {
		if err := writer.WriteField("model", g.Model); err != nil {
			return "", fmt.Errorf("failed to write model field in form data: %w", err)
		}
	}

	if g.OutputFormat != "" {
		if err := writer.WriteField("output_format", g.OutputFormat); err != nil {
			return "", fmt.Errorf("failed to write output_format field in form data: %w", err)
		}
	}

	if g.NegativePrompt != "" {
		if err := writer.WriteField("negative_prompt", g.NegativePrompt); err != nil {
			return "", fmt.Errorf("failed to write negative_prompt field in form data: %w", err)
		}
	}

	if g.Strength != 0 {
		if err := writer.WriteField("strength", strconv.FormatFloat(float64(g.Strength), 'f', 2, 32)); err != nil {
			return "", fmt.Errorf("failed to write strength field in this form data: %w", err)
		}
	}

	imageWriter, err := writer.CreateFormField("image")
	if err != nil {
		return "", fmt.Errorf("failed to create form field for image: %w", err)
	}

	_, err = io.Copy(imageWriter, g.Image)
	if err != nil {
		return "", fmt.Errorf("failed to copy image to form fields for request: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close form data writer: %w", err)
	}

	return writer.FormDataContentType(), nil
}

func validateAspectRatio(ratio string) error {
	parts := strings.Split(ratio, ":")

	if len(parts) != 2 {
		return errors.New("invalid number of semi-colons in aspect ratio")
	}

	if _, err := strconv.Atoi(parts[0]); err != nil {
		return fmt.Errorf("width ratio is not an integer: %w", err)
	}

	if _, err := strconv.Atoi(parts[1]); err != nil {
		return fmt.Errorf("height ratio is not an integer: %w", err)
	}

	return nil
}

func (g Generate3Request) Validate() error {
	if g.Prompt == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	if len(g.Prompt) > 10000 {
		return fmt.Errorf("prompt of length %d is too long; must be 10,000 characters or less", len(g.Prompt))
	}

	if g.Model != "sd3" && g.Model != "sd3turbo" {
		return fmt.Errorf("model %q is invalid; must be either \"sd3\" or \"sd3turbo\"", g.Model)
	}

	if g.AspectRatio != "" {
		if err := validateAspectRatio(g.AspectRatio); err != nil {
			return fmt.Errorf("invalid aspect ratio %q: %w", g.AspectRatio, err)
		}
	}

	return nil
}

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
