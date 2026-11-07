package logic

import (
	"math"
)

type FrequencyResult struct {
	Note100  string // the note at or below the frequency
	Cents100 int    // 0 to 99 cents above Note100
	Note50   string // the nearest note
	Cents50  int    // -50 to +50 cents from Note50
	Valid    bool   // the frequency is finite and positive
}

// GetFrequencyForNote returns the 12-TET frequency (A4 = 440 Hz under C4 naming) of a note name
// such as "C#" and octave in the given convention, or 0 for an unknown name
func GetFrequencyForNote(noteName string, octave int, octaveOffset int) float64 {
	for i, n := range noteNames {
		if n == noteName {
			midiNote := (octave+octaveOffset)*12 + i
			return 440.0 * math.Pow(2.0, (float64(midiNote)-69.0)/12.0)
		}
	}
	return 0.0
}

// FrequencyToNote converts a frequency to its note and cent offset against 12-TET with
// A4 = 440 Hz (MIDI note 69). octaveOffset is OctaveOffsetC3 or OctaveOffsetC4. Any finite
// positive frequency is named, including notes outside the MIDI range.
func FrequencyToNote(frequency float64, octaveOffset int) FrequencyResult {
	if !(frequency > 0) || math.IsInf(frequency, 1) {
		return FrequencyResult{}
	}
	split := splitSemitones(69.0 + 12.0*math.Log2(frequency/440.0))
	return FrequencyResult{
		Note100:  NoteName(split.semitones, octaveOffset),
		Cents100: split.cents,
		Note50:   NoteName(split.nearest, octaveOffset),
		Cents50:  split.cents50,
		Valid:    true,
	}
}
