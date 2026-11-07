# Note to Frequency

The frequency of every note, in standard tuning or in one of 45 other tunings.

## When to use it

- Tuning a synthesizer, a drum or an 808 kick to the key of a song.
- Finding the frequency of a note for an EQ, a filter or a resonance.
- Exploring historic temperaments, just intonation or scales from other musical traditions.

## How it works

In standard tuning (12-TET, twelve-tone equal temperament) every semitone is the same step: the frequency rises by the factor 2^(1/12), about 1.0595, and doubles every octave. One note is fixed as the **reference**, usually A = 440 Hz, and all others follow from it.

Other tunings divide the octave differently. The tunings come from Scala scale files built into the app; a scale file only lists its steps above an unnamed first step. The **scale root** decides which note that first step is on. Each tuning starts on its own tonic, C for most of them (D for Bayati, E for Huzam), and you can choose another root; the reference note keeps its frequency either way. With Kirnberger III on C, for example, C–E is a pure third; on A, the pure third is A–C#. For tunings with 12 steps, the table shows how far the tuning places every note from equal temperament in **cents** (100 cents = 1 semitone).

Tunings with another number of steps (such as 5, 7 or 13) do not fit the twelve note names. They are laid out key by key, as in Scala: step 0 is on the root key at or below the reference, and each next MIDI key plays the next step. In a 7-step scale the key "A3" is just a key, not necessarily the musical A.

## How to use it

1. **Set the reference**: the reference note and its frequency. The default is A3 = 440 Hz (in the C3 convention, where middle C is C3). Write the note with a sharp if needed and its octave, e.g. A3, C#4 or C-2; the frequency can be 1 Hz to 20 kHz.
2. **Choose the middle C convention**, C3 or C4. It only changes the octave numbers in the note names.
3. **Choose a tuning** (default: Equal · 12-TET). The 46 tunings are grouped into **Equal**, **Just**, **Historic** and **World**; the selected one is check-marked.
4. **Choose the scale root**, C to B. The tuning's own tonic is marked "scale default"; choosing another tuning puts the root back on its tonic.
5. **Read the table.**

## Reading the results

- 132 notes from C-2 to B8 (C-1 to B9 in the C4 convention), each with its **frequency** in Hz.
- The **cents** column shows how far the tuning places the note from equal temperament (0 in 12-TET).
- For tunings without 12 steps, the first column is **Step**: the step, and in brackets the period (usually the octave) counted from the root key at or below the reference, e.g. "2 (1)". The cents column then shows the step's interval above the root of its period.
- The **MIDI** column shows two note numbers: the usual one, where middle C is 60, and the one some devices use, where middle C is 72. MIDI note numbers end at 127, so the highest notes show "-".
- The table opens scrolled to the reference note, which is highlighted. The roots (step 0) are bold in the first column; with the defaults, that is every C.
- While the reference note or frequency cannot be used (e.g. "A" without an octave, or 0 Hz), the frequency and cents columns show "-" until you correct it, and so does the Step column, which depends on the reference key.

## In practice

- **Tuning a kick or 808** to the song: find the song's root note, read its frequency in the octave that suits the kick (often 40–60 Hz), and tune the kick's pitch to it.
- **Concert pitch** other than 440 Hz (e.g. 442 Hz for some orchestras, 432 Hz): change the reference frequency; the whole table follows.
- **Reset** returns to the defaults.

## Examples

- In standard tuning, A3 (C3 convention) is 440 Hz, and C3 (middle C) is 261.63 Hz.
- One octave down, A2 is 220 Hz.
- Historic · Kirnberger III on C, A3 = 440 Hz: C3 is 263.18 Hz (+10.26 cents) and E3 is 328.98 Hz, exactly 5/4 above C3.
- Just · Ptolemy/Zarlino (7 steps) on C, A3 = 440 Hz: the A3 key is step 2 (1) at +386.31 cents, and the key C3 (MIDI 60) is the root, step 0 (0), at 176.00 Hz.

## Good to know

- The C3 and C4 conventions name the same notes differently: what is A3 in one is A4 in the other. Manufacturers differ (e.g. Yamaha uses C3 for middle C), so pick the one your gear uses.
- **Frequency to Note** does the opposite: it finds the note for a frequency.
