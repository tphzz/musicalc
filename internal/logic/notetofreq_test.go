package logic

import (
	"math"
	"testing"

	sclres "musicalc/internal/logic/scl"
)

// TuningTestCase represents a test case with reference values from MusicMath. These values put
// each scale's 1/1 on the reference note, so Root is the reference note's pitch class here.
type TuningTestCase struct {
	TestName       string
	TuningName     string
	RefNoteName    string
	Root           int
	RefMidi        int
	RefFreq        float64
	OctaveOffset   int // 2 for C3 convention, 1 for C4 convention
	ExpectedValues []ExpectedNote
}

// ExpectedNote represents expected frequency and cents for a specific note
type ExpectedNote struct {
	NoteName      string
	Midi          int
	Frequency     float64
	Cents         float64
	FreqTolerance float64 // Allowed frequency deviation in Hz
	CentTolerance float64 // Allowed cents deviation
}

// TestTuningAccuracy tests tuning calculations against MusicMath reference values
func TestTuningAccuracy(t *testing.T) {
	testCases := []TuningTestCase{
		{
			TestName:     "Kepler Monochord No.1 with G2=500Hz",
			TuningName:   "Kepler's Monochord no.1, Harmonices Mundi (1619)",
			RefNoteName:  "G2",
			Root:         7,
			RefMidi:      55, // G2 in C3 convention: (2+2)*12 + 7 = 55
			RefFreq:      500.0,
			OctaveOffset: 2,
			ExpectedValues: []ExpectedNote{
				{NoteName: "C-2", Midi: 0, Frequency: 20.83, Cents: -2.00, FreqTolerance: 0.01, CentTolerance: 0.5},
				{NoteName: "C#-2", Midi: 1, Frequency: 21.97, Cents: -9.80, FreqTolerance: 0.01, CentTolerance: 0.5},
				{NoteName: "D-2", Midi: 2, Frequency: 23.44, Cents: +2.00, FreqTolerance: 0.01, CentTolerance: 0.5},
				{NoteName: "D#-2", Midi: 3, Frequency: 24.72, Cents: -5.90, FreqTolerance: 0.01, CentTolerance: 0.5},
				{NoteName: "E-2", Midi: 4, Frequency: 26.37, Cents: +5.90, FreqTolerance: 0.01, CentTolerance: 0.5},
				{NoteName: "F-2", Midi: 5, Frequency: 28.13, Cents: +17.60, FreqTolerance: 0.01, CentTolerance: 0.5},
				{NoteName: "F#-2", Midi: 6, Frequency: 29.30, Cents: -11.70, FreqTolerance: 0.01, CentTolerance: 0.5},
				{NoteName: "G-2", Midi: 7, Frequency: 31.25, Cents: 0.00, FreqTolerance: 0.01, CentTolerance: 0.5},
				{NoteName: "G#-2", Midi: 8, Frequency: 32.96, Cents: -7.80, FreqTolerance: 0.01, CentTolerance: 0.5},
				{NoteName: "A-2", Midi: 9, Frequency: 35.16, Cents: +3.90, FreqTolerance: 0.01, CentTolerance: 0.5},
				{NoteName: "A#-2", Midi: 10, Frequency: 37.50, Cents: +15.60, FreqTolerance: 0.01, CentTolerance: 0.5},
				{NoteName: "B-2", Midi: 11, Frequency: 39.06, Cents: -13.70, FreqTolerance: 0.01, CentTolerance: 0.5},
				// Test G2 reference itself
				{NoteName: "G2", Midi: 55, Frequency: 500.00, Cents: 0.00, FreqTolerance: 0.01, CentTolerance: 0.5},
			},
		},
		{
			TestName:     "12-tone Equal Temperament with A3=440Hz",
			TuningName:   "12 tone equal temperament",
			RefNoteName:  "A3",
			Root:         9,
			RefMidi:      69,
			RefFreq:      440.0,
			OctaveOffset: 2,
			ExpectedValues: []ExpectedNote{
				{NoteName: "A3", Midi: 69, Frequency: 440.00, Cents: 0.00, FreqTolerance: 0.01, CentTolerance: 0.5},
				{NoteName: "C3", Midi: 60, Frequency: 261.63, Cents: 0.00, FreqTolerance: 0.01, CentTolerance: 0.5},
				{NoteName: "C4", Midi: 72, Frequency: 523.25, Cents: 0.00, FreqTolerance: 0.01, CentTolerance: 0.5},
			},
		},
		{
			TestName:     "12-tone Pythagorean with A3=440Hz",
			TuningName:   "12-tone Pythagorean scale",
			RefNoteName:  "A3",
			Root:         9,
			RefMidi:      69,
			RefFreq:      440.0,
			OctaveOffset: 2,
			ExpectedValues: []ExpectedNote{
				{NoteName: "A3", Midi: 69, Frequency: 440.00, Cents: 0.00, FreqTolerance: 0.1, CentTolerance: 0.5},
				{NoteName: "C3", Midi: 60, Frequency: 260.74, Cents: -5.90, FreqTolerance: 0.1, CentTolerance: 0.5},
				{NoteName: "G3", Midi: 67, Frequency: 391.11, Cents: -3.90, FreqTolerance: 0.1, CentTolerance: 0.5},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			d := mustDiapason(t, tc.TuningName, tc.Root, tc.RefMidi, tc.RefFreq)
			for _, expected := range tc.ExpectedValues {
				result := d.Note(expected.Midi)

				// Check frequency
				freqDiff := math.Abs(result.Frequency - expected.Frequency)
				if freqDiff > expected.FreqTolerance {
					t.Errorf("%s (MIDI %d): Frequency mismatch\n  Expected: %.2f Hz\n  Got:      %.2f Hz\n  Diff:     %.2f Hz (tolerance: %.2f Hz)",
						expected.NoteName, expected.Midi,
						expected.Frequency, result.Frequency,
						freqDiff, expected.FreqTolerance)
				}

				// Check cents
				centDiff := math.Abs(result.Cents - expected.Cents)
				if centDiff > expected.CentTolerance {
					t.Errorf("%s (MIDI %d): Cents mismatch\n  Expected: %+.2f cents\n  Got:      %+.2f cents\n  Diff:     %.2f cents (tolerance: %.2f cents)",
						expected.NoteName, expected.Midi,
						expected.Cents, result.Cents,
						centDiff, expected.CentTolerance)
				}

				// Log success for debugging
				if freqDiff <= expected.FreqTolerance && centDiff <= expected.CentTolerance {
					t.Logf("✓ %s (MIDI %d): %.2f Hz (%+.2f cents) - PASS",
						expected.NoteName, expected.Midi,
						result.Frequency, result.Cents)
				}
			}
		})
	}
}

// TestReferenceNoteAccuracy specifically tests that reference notes are tuned correctly
func TestReferenceNoteAccuracy(t *testing.T) {
	testCases := []struct {
		name          string
		tuningName    string
		refMidi       int
		refFreq       float64
		freqTolerance float64
	}{
		{"A4=440Hz in 12-TET", "12 tone equal temperament", 69, 440.0, 0.01},
		{"G2=500Hz in Kepler", "Kepler's Monochord no.1, Harmonices Mundi (1619)", 55, 500.0, 0.01},
		{"C4=261.63Hz in Pythagorean", "12-tone Pythagorean scale", 60, 261.63, 0.01},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mustDiapason(t, tc.tuningName, tc.refMidi%12, tc.refMidi, tc.refFreq).Note(tc.refMidi)

			freqDiff := math.Abs(result.Frequency - tc.refFreq)
			if freqDiff > tc.freqTolerance {
				t.Errorf("Reference note not tuned correctly\n  Expected: %.2f Hz\n  Got:      %.2f Hz\n  Diff:     %.4f Hz",
					tc.refFreq, result.Frequency, freqDiff)
			} else {
				t.Logf("✓ Reference note correctly tuned: %.2f Hz", result.Frequency)
			}
		})
	}
}

// Scale names used by the root and degree tests
const (
	tet12         = "12 tone equal temperament"
	kirnberger3   = "Kirnberger 3: 1/4 synt. comma (1744)"
	werckmeister3 = "Harmonic equal-beating version of the famous well-temperament (2006)"
	kepler1       = "Kepler's Monochord no.1, Harmonices Mundi (1619)"
	pythagorean   = "12-tone Pythagorean scale"
	ptolemy       = "Ptolemy's Intense Diatonic Syntonon, also Zarlino's scale"
	bohlenPierce  = "See Bohlen, H. 13-Tonstufen in der Duodezime, Acustica 39: 76-86 (1978)"
	huzam         = "Arabic Huzam with perde segah on E by Dr. Oz."
)

func mustDiapason(t *testing.T, name string, root, refMidi int, refFreq float64) *Diapason {
	t.Helper()
	d, err := NewDiapason(name, root, refMidi, refFreq)
	if err != nil {
		t.Fatalf("NewDiapason(%q, %d, %d, %v): %v", name, root, refMidi, refFreq, err)
	}
	return d
}

// TestScaleRoot checks that a scale's 1/1 goes on the chosen root while the reference key keeps
// its frequency. C-rooted historic temperaments only show their defining intervals on C.
func TestScaleRoot(t *testing.T) {
	const c3, e3, g3, a3, c4 = 60, 64, 67, 69, 72 // C3 naming
	k := mustDiapason(t, kirnberger3, 0, a3, 440)
	for _, tc := range []struct {
		midi int
		hz   float64
	}{{c3, 263.1814}, {e3, 328.9767}, {g3, 393.5480}, {c4, 526.3628}, {a3, 440}} {
		if got := k.Note(tc.midi).Frequency; !near(got, tc.hz, 5e-4) {
			t.Errorf("Kirnberger III on C: MIDI %d = %.4f Hz, want %.4f", tc.midi, got, tc.hz)
		}
	}
	if r := k.Note(e3).Frequency / k.Note(c3).Frequency; !near(r, 5.0/4, 1e-12) {
		t.Errorf("Kirnberger III on C: E/C = %.15f, want the pure third 5/4", r)
	}
	if got := k.Note(c3).Cents; !near(got, 10.26, 0.01) {
		t.Errorf("Kirnberger III on C: C3 is %+.2f cents from 12-TET, want +10.26", got)
	}
	// On A, the root the reference note gives, the same pattern is transposed to A
	if got := mustDiapason(t, kirnberger3, 9, a3, 440).Note(c3).Frequency; !near(got, 260.7407, 5e-4) {
		t.Errorf("Kirnberger III on A: C3 = %.4f Hz, want 260.7407", got)
	}

	w := mustDiapason(t, werckmeister3, 0, a3, 440)
	if got := w.Note(c3).Frequency; !near(got, 263.3978, 5e-4) {
		t.Errorf("Werckmeister III on C: C3 = %.4f Hz, want 263.3978", got)
	}
	if r := w.Note(g3).Frequency / w.Note(c3).Frequency; !near(r, 9808.0/6561, 1e-12) {
		t.Errorf("Werckmeister III on C: G/C = %.15f, want 9808/6561", r)
	}

	// Kepler with the reference G2 = 500 Hz: on C its low D# and E move
	kc := mustDiapason(t, kepler1, 0, 55, 500)
	for _, tc := range []struct {
		midi int
		hz   float64
	}{{3, 25.00}, {4, 26.04}, {55, 500}} {
		if got := kc.Note(tc.midi).Frequency; !near(got, tc.hz, 0.01) {
			t.Errorf("Kepler on C: MIDI %d = %.2f Hz, want %.2f", tc.midi, got, tc.hz)
		}
	}

	// Pythagorean C and G are the same on C and on A: the fifths around them are pure either way
	for _, root := range []int{0, 9} {
		p := mustDiapason(t, pythagorean, root, a3, 440)
		if c, g := p.Note(c3).Frequency, p.Note(g3).Frequency; !near(c, 260.74, 0.01) || !near(g, 391.11, 0.01) {
			t.Errorf("Pythagorean on %d: C3 %.2f, G3 %.2f Hz, want 260.74 and 391.11", root, c, g)
		}
	}

	// Huzam has 12 steps but a 441/220 period
	h := mustDiapason(t, huzam, 4, a3, 440)
	if e, e4 := h.Note(e3).Frequency, h.Note(76).Frequency; !near(e, 329.2517, 5e-4) || !near(e4/e, 441.0/220, 1e-12) {
		t.Errorf("Huzam on E: E3 %.4f Hz, E4/E3 %.12f, want 329.2517 and 441/220", e, e4/e)
	}
}

// TestEqualTemperamentIgnoresRoot checks that rotating an equal temperament changes nothing
func TestEqualTemperamentIgnoresRoot(t *testing.T) {
	for root := 0; root < 12; root++ {
		d := mustDiapason(t, tet12, root, 69, 440)
		for midi := 0; midi < 132; midi++ {
			n := d.Note(midi)
			want := 440 * math.Pow(2, float64(midi-69)/12)
			if !near(n.Frequency, want, want*1e-9) || !near(n.Cents, 0, 1e-6) {
				t.Fatalf("12-TET on %d: MIDI %d = %v Hz %+v cents, want %v Hz", root, midi, n.Frequency, n.Cents, want)
			}
		}
	}
}

// TestScaleDegrees checks degree, period and frequency of scales without 12 steps, which follow the
// keys degree by degree from the root
func TestScaleDegrees(t *testing.T) {
	type key struct {
		midi, degree, period int
		hz, degreeCents      float64
	}
	for _, tc := range []struct {
		name     string
		root     int
		rootMidi int
		steps    int
		keys     []key
	}{
		{ptolemy, 0, 60, 7, []key{
			{60, 0, 0, 176, 0},
			{61, 1, 0, 198, 203.91},
			{67, 0, 1, 352, 0},
			{69, 2, 1, 440, 386.31},
			{59, 6, -1, 165, 1088.27},
		}},
		{bohlenPierce, 0, 60, 13, []key{
			{60, 0, 0, 205.3333, 0},
			{69, 9, 0, 440, 1319.44},
			{72, 12, 0, 570.3704, 1768.72},
			{73, 0, 1, 616, 0},
		}},
	} {
		d := mustDiapason(t, tc.name, tc.root, 69, 440)
		if d.RootMidi() != tc.rootMidi {
			t.Errorf("%s: root key %d, want %d", tc.name, d.RootMidi(), tc.rootMidi)
		}
		if n, err := ScaleSteps(tc.name); err != nil || n != tc.steps {
			t.Errorf("%s: %d steps (%v), want %d", tc.name, n, err, tc.steps)
		}
		for _, k := range tc.keys {
			n := d.Note(k.midi)
			if n.Degree != k.degree || n.Period != k.period || !near(n.Frequency, k.hz, 5e-4) || !near(n.DegreeCents, k.degreeCents, 0.01) {
				t.Errorf("%s MIDI %d: degree %d period %d, %.4f Hz, %.2f cents; want %d, %d, %.4f Hz, %.2f cents",
					tc.name, k.midi, n.Degree, n.Period, n.Frequency, n.DegreeCents, k.degree, k.period, k.hz, k.degreeCents)
			}
		}
	}
}

// TestReferenceMidiZero checks that the lowest MIDI note is a valid reference, not "missing"
func TestReferenceMidiZero(t *testing.T) {
	d := mustDiapason(t, tet12, 0, 0, 10)
	for _, tc := range []struct {
		midi int
		hz   float64
	}{{0, 10}, {12, 20}, {69, 538.1737}} {
		if got := d.Note(tc.midi).Frequency; !near(got, tc.hz, 5e-4) {
			t.Errorf("12-TET with MIDI 0 = 10 Hz: MIDI %d = %.4f Hz, want %.4f", tc.midi, got, tc.hz)
		}
	}
	// The root key may not lie below MIDI 0: D is then the one above the reference
	p := mustDiapason(t, ptolemy, 2, 0, 100)
	if p.RootMidi() != 2 || !near(p.Note(0).Frequency, 100, 1e-9) || !near(p.Note(2).Frequency, 120, 1e-9) {
		t.Errorf("Ptolemy on D with MIDI 0 = 100 Hz: root key %d, MIDI 0 %.4f, MIDI 2 %.4f Hz; want 2, 100, 120",
			p.RootMidi(), p.Note(0).Frequency, p.Note(2).Frequency)
	}
}

// TestEveryScaleKeepsReference checks every bundled scale and root against an independent
// periodic formula, with the reference key exact and every value finite and positive
func TestEveryScaleKeepsReference(t *testing.T) {
	const refHz = 432.123456
	for name := range sclres.AvailableScales {
		for root := 0; root < 12; root++ {
			for _, ref := range []int{0, 57, 69, 127} {
				d := mustDiapason(t, name, root, ref, refHz)
				if got := d.Note(ref).Frequency; !near(got, refHz, refHz*1e-9) {
					t.Fatalf("%s, root %d: reference MIDI %d = %v Hz, want %v", name, root, ref, got, refHz)
				}
				periodCents := d.scale.Tones[d.scale.Count-1].Cents
				rootHz := d.Note(d.RootMidi()).Frequency
				for midi := 0; midi < 132; midi++ {
					n := d.Note(midi)
					want := rootHz * math.Pow(2, (float64(n.Period)*periodCents+n.DegreeCents)/1200)
					if !(n.Frequency > 0) || math.IsInf(n.Frequency, 0) || !near(n.Frequency, want, want*1e-9) {
						t.Fatalf("%s, root %d, ref %d: MIDI %d = %v Hz, want %v", name, root, ref, midi, n.Frequency, want)
					}
				}
			}
		}
	}
}

func TestNewDiapasonRejectsInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		name          string
		root, refMidi int
		refHz         float64
	}{
		{"no such tuning", 0, 69, 440},
		{tet12, -1, 69, 440},
		{tet12, 12, 69, 440},
		{tet12, 0, -1, 440},
		{tet12, 0, 128, 440},
		{tet12, 0, 69, 0},
		{tet12, 0, 69, -1},
		{tet12, 0, 69, 0.5},
		{tet12, 0, 69, 20001},
		{tet12, 0, 69, math.NaN()},
		{tet12, 0, 69, math.Inf(1)},
	} {
		if _, err := NewDiapason(tc.name, tc.root, tc.refMidi, tc.refHz); err == nil {
			t.Errorf("NewDiapason(%q, %d, %d, %v) succeeded, want an error", tc.name, tc.root, tc.refMidi, tc.refHz)
		}
	}
}

// TestScaleStepsOfEveryTuning checks the step count that chooses between note names and degrees:
// every bundled tuning has one, and exactly 27 of the 46 have 12 steps
func TestScaleStepsOfEveryTuning(t *testing.T) {
	twelve := 0
	for name := range sclres.AvailableScales {
		n, err := ScaleSteps(name)
		if err != nil || n < 1 {
			t.Errorf("%s: %d steps, %v", name, n, err)
		}
		if n == 12 {
			twelve++
		}
	}
	if twelve != 27 || len(sclres.AvailableScales) != 46 {
		t.Errorf("%d of %d tunings have 12 steps, want 27 of 46", twelve, len(sclres.AvailableScales))
	}
	if _, err := ScaleSteps("no such tuning"); err == nil {
		t.Error("an unknown tuning has a step count")
	}
}
