package stability_test

import (
	"testing"

	"github.com/SethCurry/sdcli/pkg/stability"
)

func TestAspectRatioString(t *testing.T) {
	testCases := []struct {
		name     string
		ratio    stability.AspectRatio
		expected string
	}{
		{
			"Simple 1:1",
			stability.AspectRatio{1, 1},
			"1:1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if output := tc.ratio.String(); output != tc.expected {
				t.Errorf("got %s want %s", output, tc.expected)
			}
		})
	}
}
