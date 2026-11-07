package logic

import (
	"errors"
	"strconv"
	"strings"
)

// noteNames are the twelve pitch classes from C, written with sharps
var noteNames = [12]string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

// Octave offsets of the two naming conventions: MIDI note 60 is C3 or C4
const (
	OctaveOffsetC3 = 2
	OctaveOffsetC4 = 1
)

// NoteName names a note number, e.g. 69 → "A3" with OctaveOffsetC3. It also names notes outside
// the MIDI range (negative or above 127), such as those of very low or high frequencies.
func NoteName(note, octaveOffset int) string {
	return noteNames[floorMod(note, 12)] + strconv.Itoa(floorDiv(note, 12)-octaveOffset)
}

// PitchClassName names pitch class pc (0 = C … 11 = B) without an octave, e.g. 1 → "C#"
func PitchClassName(pc int) string {
	return noteNames[floorMod(pc, 12)]
}

// ParseNoteName reads a note such as "A3", "c#4" or "C-2" in the given octave convention and
// returns its MIDI note number (0 to 127). Flats and other spellings are not accepted.
func ParseNoteName(s string, octaveOffset int) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty note name")
	}
	name, rest := strings.ToUpper(s[:1]), s[1:]
	if strings.HasPrefix(rest, "#") {
		name, rest = name+"#", rest[1:]
	}
	index := -1
	for i, n := range noteNames {
		if n == name {
			index = i
		}
	}
	if index < 0 {
		return 0, errors.New("unknown note name")
	}
	octave, err := strconv.Atoi(rest)
	if err != nil || strings.HasPrefix(rest, "+") {
		return 0, errors.New("missing or invalid octave")
	}
	if octave < -10 || octave > 20 {
		return 0, errors.New("note outside the MIDI range")
	}
	midi := (octave+octaveOffset)*12 + index
	if midi < 0 || midi > 127 {
		return 0, errors.New("note outside the MIDI range")
	}
	return midi, nil
}
