package logic

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

// AlignmentMic is one microphone and the delay it needs to line up with the reference mic.
// A mic farther from the source than the reference gets a negative delay: it would have to be
// moved earlier, which a plain delay can't do.
type AlignmentMic struct {
	Name              string
	DistanceMeters    float64
	DelayMS           float64
	DelaySamples      int
	IsBeyondReference bool // farther from the source than the reference; the delay is negative
	IsReference       bool // at the reference distance; no delay needed
}

// referenceTolerance is how close (in metres) a mic must be to the reference distance to count as the reference
const referenceTolerance = 1e-9

// AlignmentDelayCalculator works out the delays that line close mics up with the reference mic.
// All distances are measured from the same sound source, named in Source.
type AlignmentDelayCalculator struct {
	Source                  string
	TemperatureC            float64
	SampleRate              int
	ReferenceDistanceMeters float64
	Mics                    []AlignmentMic
}

func NewAlignmentDelayCalculator() *AlignmentDelayCalculator {
	return &AlignmentDelayCalculator{SampleRate: 48000}
}

func (c *AlignmentDelayCalculator) SetTemperature(value float64, unit string) {
	c.TemperatureC = ToCelsius(value, unit)
}

func (c *AlignmentDelayCalculator) SetSampleRateLabel(label string) {
	c.SampleRate = ParseSampleRate(label)
	if c.SampleRate <= 0 {
		c.SampleRate = 48000
	}
}

func (c *AlignmentDelayCalculator) SetReferenceDistance(value float64, unit string) {
	c.ReferenceDistanceMeters = ToMeters(value, unit)
}

func (c *AlignmentDelayCalculator) AddMic(name string, distance float64, unit string) {
	c.Mics = append(c.Mics, AlignmentMic{Name: name, DistanceMeters: ToMeters(distance, unit)})
}

func (c *AlignmentDelayCalculator) RemoveMicAt(index int) {
	if index < 0 || index >= len(c.Mics) {
		return
	}
	c.Mics = append(c.Mics[:index], c.Mics[index+1:]...)
}

func (c *AlignmentDelayCalculator) SpeedOfSoundMetersPerSecond() float64 {
	return SpeedOfSound(c.TemperatureC)
}

// Reasons a setup cannot be calculated
var (
	ErrTemperatureRange = errors.New("temperature outside the accepted range")
	ErrReferenceInvalid = errors.New("the reference distance is not a finite distance of 0 or more")
	ErrMicInvalid       = errors.New("a mic distance is not a finite distance of 0 or more")
	ErrSampleRate       = errors.New("the sample rate is not positive")
	ErrDelayTooLarge    = errors.New("a delay is too long to count in samples")
)

// maxDelaySamples bounds sample counts well inside the exactly representable integers
const maxDelaySamples = 1e15

// SampleCountInRange reports whether a delay of seconds at rate samples per second can be counted
// in whole samples exactly (below 1e15), so it is never shown as Inf or a meaningless number
func SampleCountInRange(seconds float64, rate int) bool {
	return math.Abs(seconds*float64(rate)) < maxDelaySamples
}

// Validate reports why the setup cannot be calculated, or nil
func (c *AlignmentDelayCalculator) Validate() error {
	switch {
	case !TemperatureInRange(c.TemperatureC):
		return ErrTemperatureRange
	case !validDistance(c.ReferenceDistanceMeters):
		return ErrReferenceInvalid
	case c.SampleRate <= 0:
		return ErrSampleRate
	}
	for _, m := range c.Mics {
		if !validDistance(m.DistanceMeters) {
			return ErrMicInvalid
		}
		if !SampleCountInRange((c.ReferenceDistanceMeters-m.DistanceMeters)/c.SpeedOfSoundMetersPerSecond(), c.SampleRate) {
			return ErrDelayTooLarge
		}
	}
	return nil
}

// Recalculate works out every mic's delay. For an invalid setup it clears the delays and returns
// the reason, rather than results from impossible values.
func (c *AlignmentDelayCalculator) Recalculate() error {
	if err := c.Validate(); err != nil {
		c.ClearDelays()
		return err
	}
	for i := range c.Mics {
		delayMS, delaySamples, beyond := c.DelayForDistanceMeters(c.Mics[i].DistanceMeters)
		c.Mics[i].DelayMS = delayMS
		c.Mics[i].DelaySamples = delaySamples
		c.Mics[i].IsBeyondReference = beyond
		c.Mics[i].IsReference = math.Abs(c.ReferenceDistanceMeters-c.Mics[i].DistanceMeters) <= referenceTolerance
	}
	return nil
}

// ClearDelays removes every mic's calculated delay, e.g. while the setup cannot be calculated
func (c *AlignmentDelayCalculator) ClearDelays() {
	for i := range c.Mics {
		c.Mics[i].DelayMS, c.Mics[i].DelaySamples = 0, 0
		c.Mics[i].IsBeyondReference, c.Mics[i].IsReference = false, false
	}
}

// DelayForDistanceMeters returns the delay that lines a mic at targetMeters up with the reference,
// in milliseconds and samples, and whether the mic is farther than the reference. For a farther
// mic the delay is negative: it arrives that much later than the reference.
func (c *AlignmentDelayCalculator) DelayForDistanceMeters(targetMeters float64) (float64, int, bool) {
	distanceDiff := c.ReferenceDistanceMeters - targetMeters
	beyond := distanceDiff < -referenceTolerance

	speed := c.SpeedOfSoundMetersPerSecond() // positive within the accepted temperatures

	delayMS := (distanceDiff / speed) * 1000
	delaySamples := int(math.Round(delayMS * float64(c.SampleRate) / 1000))

	return delayMS, delaySamples, beyond
}

func ToMeters(distance float64, unit string) float64 {
	if strings.EqualFold(unit, "ft") {
		return distance * 0.3048
	}
	return distance
}

func FromMeters(distanceMeters float64, unit string) float64 {
	if strings.EqualFold(unit, "ft") {
		return distanceMeters / 0.3048
	}
	return distanceMeters
}

func ParseSampleRate(rateStr string) int {
	rateStr = strings.TrimSpace(rateStr)
	if rateStr == "" {
		return 48000
	}

	s := strings.ToLower(strings.ReplaceAll(rateStr, " ", ""))

	if strings.HasSuffix(s, "khz") {
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "khz"), 64)
		if err == nil {
			return int(math.Round(v * 1000))
		}
	}
	if strings.HasSuffix(s, "hz") {
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "hz"), 64)
		if err == nil {
			return int(math.Round(v))
		}
	}

	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return int(math.Round(v))
	}

	return 48000
}
