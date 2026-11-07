package logic

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strings"

	sclres "musicalc/internal/logic/scl"

	scala "github.com/chinenual/go-scala"
)

// The supported reference frequency range. go-scala writes the reference into a keyboard mapping
// with six decimals, so a far smaller one would round to zero; above 20 kHz it is not audible.
const (
	MinReferenceHz = 1.0
	MaxReferenceHz = 20000.0
)

// NoteFrequency is one key of a Diapason
type NoteFrequency struct {
	Frequency   float64 // Hz
	Cents       float64 // against 12-TET through the reference key (positive = sharp)
	Degree      int     // scale degree, 0 = the root
	Period      int     // periods (usually octaves) above the root key, negative below
	DegreeCents float64 // the degree's interval above the root, from the .scl file
}

// Diapason lays a tuning out on the MIDI keys: scale degree 0 on the root key and the reference
// key tuned to the reference frequency, to within 1e-6 Hz as go-scala keeps six decimals. Keys
// follow one another degree by degree, as in Scala's standard keyboard mapping, so a scale without
// 12 steps does not follow the note names.
type Diapason struct {
	scale    scala.Scale
	tuning   scala.Tuning
	rootMidi int
	refMidi  int
	refFreq  float64
}

// NewDiapason builds the tuning named tuningName (a key of AvailableScales) with its root on pitch
// class root (0 = C … 11 = B) and MIDI note refMidi tuned to refFreq Hz
func NewDiapason(tuningName string, root, refMidi int, refFreq float64) (*Diapason, error) {
	info, ok := sclres.AvailableScales[tuningName]
	switch {
	case !ok:
		return nil, fmt.Errorf("unknown tuning %q", tuningName)
	case root < 0 || root > 11:
		return nil, fmt.Errorf("root pitch class %d is not 0 to 11", root)
	case refMidi < 0 || refMidi > 127:
		return nil, fmt.Errorf("reference note %d is outside the MIDI range", refMidi)
	case !(refFreq >= MinReferenceHz && refFreq <= MaxReferenceHz):
		return nil, errors.New("reference frequency is outside 1 Hz to 20 kHz")
	}

	scale, err := readScale(info)
	if err != nil {
		return nil, err
	}
	rootMidi := rootKey(root, refMidi)
	kbm, err := scala.KeyboardMappingStartScaleOnAndTuneNoteTo(rootMidi, refMidi, refFreq)
	if err != nil {
		return nil, fmt.Errorf("mapping %s: %w", info.Filename, err)
	}
	tuning, err := scala.TuningFromSCLAndKBM(scale, kbm)
	if err != nil {
		return nil, fmt.Errorf("tuning %s: %w", info.Filename, err)
	}
	return &Diapason{scale: scale, tuning: tuning, rootMidi: rootMidi, refMidi: refMidi, refFreq: refFreq}, nil
}

// ScaleSteps is the number of degrees in one period of the tuning named tuningName. Only a 12-step
// scale follows the note names. It does not depend on a reference, so it is known while the
// reference is being corrected.
func ScaleSteps(tuningName string) (int, error) {
	info, ok := sclres.AvailableScales[tuningName]
	if !ok {
		return 0, fmt.Errorf("unknown tuning %q", tuningName)
	}
	scale, err := readScale(info)
	if err != nil {
		return 0, err
	}
	return scale.Count, nil
}

func readScale(info sclres.ScaleInfo) (scala.Scale, error) {
	scale, err := scala.ScaleFromSCLStream(bytes.NewReader(sclPitchesOnly(info.Resource.Content())))
	if err != nil {
		return scale, fmt.Errorf("reading %s: %w", info.Filename, err)
	}
	return scale, nil
}

// sclPitchesOnly drops the text the Scala format allows after a pitch value, such as
// "100.0 cents" or "2187/2048 ! name", and the indentation of comments, both of which go-scala
// would reject. The description line, the first one that is not a "!" comment, stays as it is.
func sclPitchesOnly(data []byte) []byte {
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	described := false
	for i, line := range lines {
		switch {
		case strings.HasPrefix(strings.TrimSpace(line), "!"):
			lines[i] = strings.TrimSpace(line) // go-scala only skips comments starting the line
		case !described:
			described = true
		default:
			if fields := strings.Fields(line); len(fields) > 0 {
				lines[i] = fields[0]
			}
		}
	}
	return []byte(strings.Join(lines, "\n"))
}

// rootKey is the key of pitch class root at or below refMidi, or the lowest one if that would be
// below MIDI note 0. For a 12-step scale its octave does not matter; for others it decides which
// degree the reference key is.
func rootKey(root, refMidi int) int {
	key := refMidi - floorMod(refMidi-root, 12)
	if key < 0 {
		key += 12
	}
	return key
}

// RootMidi is the key that sounds the root (degree 0) nearest the reference
func (d *Diapason) RootMidi() int { return d.rootMidi }

// Note returns the frequency and scale position of MIDI note midi (0 to 131 in the table)
func (d *Diapason) Note(midi int) NoteFrequency {
	freq := d.tuning.FrequencyForMidiNote(midi)
	tet := d.refFreq * math.Pow(2, float64(midi-d.refMidi)/12.0)
	k := midi - d.rootMidi
	period := floorDiv(k, d.scale.Count)
	degree := k - period*d.scale.Count
	var degreeCents float64
	if degree > 0 {
		degreeCents = d.scale.Tones[degree-1].Cents
	}
	return NoteFrequency{
		Frequency:   freq,
		Cents:       1200.0 * math.Log2(freq/tet),
		Degree:      degree,
		Period:      period,
		DegreeCents: degreeCents,
	}
}
