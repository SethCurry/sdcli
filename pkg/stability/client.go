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
	AspectRatio    string  `json:"aspect_ratio"`
	Prompt         string  `json:"prompt"`
	Model          string  `json:"model"`
	OutputFormat   string  `json:"output_format"`
	NegativePrompt string  `json:"negative_prompt"`
	Strength       float32 `json:"strength"`
	Image          []byte  `json:"image"`
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

type Generate3Option func(*multipart.Writer) error

func WithAspectRatio(ratio string) Generate3Option {
	return func(req *multipart.Writer) error {
		return req.WriteField("aspect_ratio", ratio)
	}
}

func WithPrompt(prompt string) Generate3Option {
	return func(req *multipart.Writer) error {
		return req.WriteField("prompt", prompt)
	}
}

func WithModel(model string) Generate3Option {
	return func(req *multipart.Writer) error {
		return req.WriteField("model", model)
	}
}

func WithOutputFormat(format string) Generate3Option {
	return func(req *multipart.Writer) error {
		return req.WriteField("output_format", format)
	}
}

func WithNegativePrompt(prompt string) Generate3Option {
	return func(req *multipart.Writer) error {
		return req.WriteField("negative_prompt", prompt)
	}
}

func WithStrength(strength float32) Generate3Option {
	return func(req *multipart.Writer) error {
		return req.WriteField("strength", strconv.FormatFloat(float64(strength), 'f', 2, 32))
	}
}

func WithImage(reader io.Reader) Generate3Option {
	return func(req *multipart.Writer) error {
		writer, err := req.CreateFormField("image")
		if err != nil {
			return fmt.Errorf("failed to create image field in request: %w", err)
		}

		_, err = io.Copy(writer, reader)

		return err
	}
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

func (c *Client) Generate3(ctx context.Context, writeTo io.Writer, options ...Generate3Option) error {
	reqURL := fmt.Sprintf("%s/v2beta/stable-image/generate/sd3", c.baseURL)

	var formBuf bytes.Buffer

	writer := multipart.NewWriter(&formBuf)

	for _, v := range options {
		err := v(writer)
		if err != nil {
			return err
		}
	}

	err := writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, &formBuf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
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
