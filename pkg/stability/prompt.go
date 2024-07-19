package stability

import (
	"errors"
	"fmt"
)

// Prompt encapsulates a user's prompt.  It is used to provide the Validate
// method to perform validation that the prompt is not too long.
type Prompt string

// Validate checks to ensure that the prompt is valid by returning an error
// if it is longer than 10,000 characters.
func (p Prompt) Validate() error {
	if len(p) > 10000 {
		return fmt.Errorf("prompt of length %d is invalid: %w", len(p), ErrPromptTooLong)
	}

	return nil
}

// ErrPromptTooLong is returned when the prompt exceeds the maximum length of 10,000 characters.
var ErrPromptTooLong = errors.New("prompt is too long, maximum length is 10,000 characters")
