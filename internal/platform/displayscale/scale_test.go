package displayscale

import (
	"math"
	"os"
	"testing"
)

func TestConfigureUsesOSScaleAndPreservesExplicitOverride(t *testing.T) {
	for _, value := range []string{"", "auto", "1.5"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("FYNE_SCALE", value)
			if err := Configure(); err != nil {
				t.Fatal(err)
			}
			want := value
			if value == "" {
				want = "auto"
			}
			if got := os.Getenv("FYNE_SCALE"); got != want {
				t.Fatalf("got %q want %q", got, want)
			}
		})
	}
}

func TestScaleChangeIgnoresInitialAndInvalidReadings(t *testing.T) {
	var scale observer
	for _, tc := range []struct {
		value   float32
		changed bool
	}{
		{0, false}, {float32(math.NaN()), false}, {float32(math.Inf(1)), false},
		{1, false}, {1, false}, {1.25, true}, {1.25, false}, {1.5, true}, {1, true},
	} {
		if got := scale.changed(tc.value); got != tc.changed {
			t.Errorf("scale %v: changed %v want %v", tc.value, got, tc.changed)
		}
	}
}
