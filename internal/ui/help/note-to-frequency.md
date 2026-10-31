# Note to Frequency

The frequency of every note, in standard tuning or in one of 46 other tunings.

## When to use it

- Tuning a synthesizer, a drum or an 808 kick to the key of a song.
- Finding the frequency of a note for an EQ, a filter or a resonance.
- Exploring historic temperaments, just intonation or scales from other musical traditions.

## How it works

In standard tuning (12-TET, twelve-tone equal temperament) every semitone is the same step: the frequency rises by the factor 2^(1/12), about 1.0595, and doubles every octave. One note is fixed as the **reference**, usually A = 440 Hz, and all others follow from it.

Other tunings divide the octave differently. The calculator loads them from Scala scale files and shows, for every note, how far it is from equal temperament in **cents** (100 cents = 1 semitone).

## How to use it

1. **Set the reference**: the reference note and its frequency. The default is A3 = 440 Hz (in the C3 convention, where middle C is C3).
2. **Choose the middle C convention**, C3 or C4. It only changes the octave numbers in the note names.
3. **Choose a tuning** (default: Equal · 12-TET). The 46 tunings are grouped into **Equal**, **Just**, **Historic** and **World**; the selected one is check-marked.
4. **Read the table.**

## Reading the results

- Every note from C-2 to B8, with its **frequency** in Hz.
- The **cents** column shows how far the tuning places the note from equal temperament (0 in 12-TET).
- The **MIDI** columns give the note number in both conventions (C3 = 60 and C4 = 60). MIDI note numbers only go from 0 to 127, so the highest notes show "-".
- The table opens scrolled to the reference note, which is highlighted. Every C is bold, marking the start of an octave.

## In practice

- **Tuning a kick or 808** to the song: find the song's root note, read its frequency in the octave that suits the kick (often 40–60 Hz), and tune the kick's pitch to it.
- **Concert pitch** other than 440 Hz (e.g. 442 Hz for some orchestras, 432 Hz): change the reference frequency; the whole table follows.
- **Reset** returns to the defaults.

## Examples

- In standard tuning, A3 (C3 convention) is 440 Hz, and C3 (middle C) is 261.63 Hz.
- One octave down, A2 is 220 Hz.

## Good to know

- The C3 and C4 conventions name the same notes differently: what is A3 in one is A4 in the other. Manufacturers differ (e.g. Yamaha uses C3 for middle C), so pick the one your gear uses.
- **Frequency to Note** does the opposite: it finds the note for a frequency.
