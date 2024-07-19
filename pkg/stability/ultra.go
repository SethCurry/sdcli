package stability

import (
	"fmt"
	"mime/multipart"
)

type GenerateUltraRequest struct {
	Prompt         Prompt
	NegativePrompt Prompt
	AspectRatio    AspectRatio
	OutputFormat   string
}

func (g GenerateUltraRequest) validate() error {
	if err := g.Prompt.validate(); err != nil {
		return err
	}

	if err := g.NegativePrompt.validate(); err != nil {
		return err
	}

	if err := g.AspectRatio.validate(); err != nil {
		return err
	}

	return nil
}

func (g GenerateUltraRequest) toFormData(writer *multipart.Writer) error {
	if err := writer.WriteField("prompt", string(g.Prompt)); err != nil {
		return fmt.Errorf("failed to write prompt field: %w", err)
	}

	if len(g.NegativePrompt) != 0 {
		if err := writer.WriteField("negative_prompt", string(g.NegativePrompt)); err != nil {
			return fmt.Errorf("failed to write negative prompt field: %w", err)
		}
	}

	if err := writer.WriteField("aspect_ratio", g.AspectRatio.String()); err != nil {
		return fmt.Errorf("failed to write aspect ratio field: %w", err)
	}

	if err := writer.WriteField("output_format", g.OutputFormat); err != nil {
		return fmt.Errorf("failed to write output format field: %w", err)
	}

	return nil
}
