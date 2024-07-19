package stability

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// AspectRatio encodes an aspect ratio such as "16:9" or "4:3".
type AspectRatio struct {
	Width  int
	Height int
}

// String converts the AspectRatio back into a colon-delimited string
// like "4:3".
func (a AspectRatio) String() string {
	return fmt.Sprintf("%d:%d", a.Width, a.Height)
}

// validate ensures that the aspect ratio is valid by checking
// - Whether height or width is less than 1
// - Whether the aspect ratio is one recognized by the Stable Diffusion API
func (a AspectRatio) validate() error {
	if a.Width <= 0 {
		return fmt.Errorf("aspect ratio width is less than 0: %d", a.Width)
	}

	if a.Height <= 0 {
		return fmt.Errorf("aspect ratio height is less than 0: %d", a.Height)
	}

	isValidRatio := false

	for _, v := range validAspectRatios {
		if v.Width == a.Width && v.Height == a.Height {
			isValidRatio = true
			break
		}
	}

	if !isValidRatio {
		return fmt.Errorf("aspect ratio is not supported: %s", a.String())
	}
	return nil
}

var validAspectRatios = []AspectRatio{
	{1, 1},
	{16, 9},
	{21, 9},
	{2, 3},
	{3, 2},
	{4, 5},
	{5, 4},
	{9, 16},
	{9, 21},
}

// ParseAspectRatio takes a colon-delimited aspect ratio like "4:3" and returns
// an AspectRatio representing it.
func ParseAspectRatio(ratio string) (*AspectRatio, error) {
	parsed := AspectRatio{}

	parts := strings.Split(ratio, ":")

	if len(parts) != 2 {
		return nil, errors.New("invalid number of semi-colons in aspect ratio")
	}

	if width, err := strconv.Atoi(parts[0]); err != nil {
		return nil, fmt.Errorf("width ratio is not an integer: %w", err)
	} else {
		parsed.Width = width
	}

	if height, err := strconv.Atoi(parts[1]); err != nil {
		return nil, fmt.Errorf("height ratio is not an integer: %w", err)
	} else {
		parsed.Height = height
	}

	return &parsed, nil
}
