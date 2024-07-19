package stability

import (
	"errors"
	"fmt"
)

// Strength represents the "strength" parameter passed to many endpoints
// on the Stability API.
type Strength float32

// Validate ensures that the strength parameter is between 0 and 1.
func (s Strength) Validate() error {
	if s < 0 || s > 1.0 {
		return fmt.Errorf("strength %f is invalid: %w", float32(s), ErrStrengthOutOfRange)
	}

	return nil
}

// ErrStrengthOutOfRange is returned when the strength parameter is not between 0.0 and 1.0.
var ErrStrengthOutOfRange = errors.New("strength must be between 0.0 and 1.0")
