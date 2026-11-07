package logic

import (
	"errors"
	"math"
	"testing"
)

// TestAlignmentValidateReasons checks every reason a setup cannot be calculated, and that the
// delays are cleared rather than left from an earlier valid setup
func TestAlignmentValidateReasons(t *testing.T) {
	valid := func() *AlignmentDelayCalculator {
		c := NewAlignmentDelayCalculator()
		c.TemperatureC, c.ReferenceDistanceMeters = 20, 3
		c.AddMic("Top", 1, UnitMeters)
		return c
	}
	for _, tc := range []struct {
		name string
		edit func(c *AlignmentDelayCalculator)
		want error
	}{
		{"temperature", func(c *AlignmentDelayCalculator) { c.TemperatureC = -600 }, ErrTemperatureRange},
		{"reference", func(c *AlignmentDelayCalculator) { c.ReferenceDistanceMeters = math.NaN() }, ErrReferenceInvalid},
		{"mic", func(c *AlignmentDelayCalculator) { c.Mics[0].DistanceMeters = -1 }, ErrMicInvalid},
		{"sample rate", func(c *AlignmentDelayCalculator) { c.SampleRate = 0 }, ErrSampleRate},
		{"too large", func(c *AlignmentDelayCalculator) { c.ReferenceDistanceMeters = 1e300 }, ErrDelayTooLarge},
	} {
		c := valid()
		if err := c.Recalculate(); err != nil || c.Mics[0].DelaySamples != 280 {
			t.Fatalf("%s: valid setup gave %v, %+v", tc.name, err, c.Mics[0])
		}
		tc.edit(c)
		if err := c.Recalculate(); !errors.Is(err, tc.want) || c.Mics[0].DelayMS != 0 || c.Mics[0].DelaySamples != 0 {
			t.Errorf("%s: %v, %+v; want %v and cleared delays", tc.name, err, c.Mics[0], tc.want)
		}
	}
	if !SampleCountInRange(1, 192000) || SampleCountInRange(1e10, 192000) || SampleCountInRange(math.NaN(), 48000) {
		t.Error("SampleCountInRange bounds")
	}
}
