package stability

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
)

// SD3Model is a representation of a Stable Diffusion 3 model.
type SD3Model string

// exists returns whether the model name is a known Stable Diffusion 3 model
// by checking whether it exists in AllSD3Models.
func (s SD3Model) exists() bool {
	for _, m := range AllSD3Models {
		if s == m {
			return true
		}
	}

	return false
}

func (s SD3Model) validate() error {
	if !s.exists() {
		return fmt.Errorf("%w: %s", ErrUnknownModel, s)
	}

	return nil
}

// ErrUnknownModel is returned when the user selects a model that the Stable Diffusion API does
// not support.
var ErrUnknownModel = errors.New("unrecognized model name")

const (
	// SD3Medium represents the sd3-medium model for Stable Diffusion 3.
	SD3Medium = SD3Model("sd3-medium")

	// SD3Large represents the sd3-large model for Stable Diffusion 3.
	SD3Large = SD3Model("sd3-large")

	// SD3LargeTurbo represents the sd3-large-turbo model for Stable Diffusion 3.
	SD3LargeTurbo = SD3Model("sd3-large-turbo")
)

// AllSD3Models stores a list of all of the valid Stable Diffusion 3 models.
var AllSD3Models = []SD3Model{
	SD3Medium,
	SD3Large,
	SD3LargeTurbo,
}

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
	Prompt Prompt

	// Model is the model to use, since there are several variants of Stable Diffusion 3.
	//
	// It is a required field.
	Model SD3Model

	// OutputFormat is the format of the image to generate.
	// Valid values are png and jpeg.
	//
	// It is a required field.
	OutputFormat string

	// NegativePrompt is the negative prompt provided to Stable Diffusion 3.
	NegativePrompt Prompt

	// Strength is the strength of the prompt.
	// It ranges from 0.0 to 1.0.
	Strength Strength

	// Image allows providing an image for image-to-image generation.
	Image io.Reader
}

// toFormData converts the Generate3Request into a form-data payload that can be sent to the Stable Diffusion 3 API.
// It returns the Content-Type header the form-data payload should be sent with, along with an error
// if it was unable to write any of the fields to the form data.
func (g Generate3Request) toFormData(writer *multipart.Writer) error {
	if err := writer.WriteField("aspect_ratio", g.AspectRatio.String()); err != nil {
		return fmt.Errorf("failed to write aspect_ratio field in form data: %w", err)
	}

	if err := writer.WriteField("prompt", string(g.Prompt)); err != nil {
		return fmt.Errorf("failed to write prompt field in form data: %w", err)
	}

	if g.Model != "" {
		if err := writer.WriteField("model", string(g.Model)); err != nil {
			return fmt.Errorf("failed to write model field in form data: %w", err)
		}
	}

	if g.OutputFormat != "" {
		if err := writer.WriteField("output_format", g.OutputFormat); err != nil {
			return fmt.Errorf("failed to write output_format field in form data: %w", err)
		}
	}

	if g.NegativePrompt != "" {
		if err := writer.WriteField("negative_prompt", string(g.NegativePrompt)); err != nil {
			return fmt.Errorf("failed to write negative_prompt field in form data: %w", err)
		}
	}

	if g.Strength != 0 {
		if err := writer.WriteField("strength", strconv.FormatFloat(float64(g.Strength), 'f', 2, 32)); err != nil {
			return fmt.Errorf("failed to write strength field in this form data: %w", err)
		}
	}

	imageWriter, err := writer.CreateFormField("image")
	if err != nil {
		return fmt.Errorf("failed to create form field for image: %w", err)
	}

	_, err = io.Copy(imageWriter, g.Image)
	if err != nil {
		return fmt.Errorf("failed to copy image to form fields for request: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close form data writer: %w", err)
	}

	return nil
}

func (g Generate3Request) validate() error {
	if g.Prompt == "" {
		return errors.New("prompt cannot be empty")
	}

	if err := g.Prompt.Validate(); err != nil {
		return fmt.Errorf("prompt is invalid: %w", err)
	}

	if err := g.NegativePrompt.Validate(); err != nil {
		return fmt.Errorf("negative prompt is invalid: %w", err)
	}

	if g.AspectRatio.Width == 0 || g.AspectRatio.Height == 0 {
		return fmt.Errorf("invalid aspect ratio: %q", g.AspectRatio.String())
	}

	if err := g.Model.validate(); err != nil {
		return err
	}

	return nil
}
