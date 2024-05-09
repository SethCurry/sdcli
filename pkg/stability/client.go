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

func Generate3(ctx context.Context, baseURL string, apiKey string, options ...Generate3Option) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/v2beta/stable-image/generate/sd3", baseURL)

	var formBuf bytes.Buffer

	writer := multipart.NewWriter(&formBuf)

	for _, v := range options {
		err := v(writer)
		if err != nil {
			return nil, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, &formBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "image/*")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image from response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("got unexpected status code %d while generating image. Response: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
