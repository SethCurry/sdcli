package stability

import (
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
)

// Generate3Request encapsulates all of the parameters for generating an image with
// the Stable Diffusion 3 API.
type Generate3Request struct {
	// AspectRatio is the aspect ratio of the image to generate.
	// This is limited to a particular set of values, arbitrary
	// ratios will not work.
	//
	// The currently accepted values are:
	// 16:9, 1:1, 21:9, 2:3, 3:2, 4:5, 5:4, 9:16, 9:21
	AspectRatio AspectRatio

	// Prompt is the prompt to use for generating an image.
	//
	// It is a required field.
	Prompt string

	// Model is the model to use, since there are several variants of Stable Diffusion 3.
	//
	// It is a required field.
	Model string

	// OutputFormat is the format of the image to generate.
	// Valid values are png and jpeg.
	//
	// It is a required field.
	OutputFormat string

	// NegativePrompt is the negative prompt provided to Stable Diffusion 3.
	NegativePrompt string

	// Strength is the strength of the prompt.
	// It ranges from 0.0 to 1.0.
	Strength float32

	// Image allows providing an image for image-to-image generation.
	Image io.Reader
}

// ToFormData converts the Generate3Request into a form-data payload that can be sent to the Stable Diffusion 3 API.
// It returns the Content-Type header the form-data payload should be sent with, along with an error
// if it was unable to write any of the fields to the form data.
func (g Generate3Request) ToFormData(writeTo io.Writer) (string, error) {
	writer := multipart.NewWriter(writeTo)

	if err := writer.WriteField("aspect_ratio", g.AspectRatio.String()); err != nil {
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

	if g.AspectRatio.Width == 0 || g.AspectRatio.Height == 0 {
		return fmt.Errorf("invalid aspect ratio: %q", g.AspectRatio.String())
	}

	return nil
}