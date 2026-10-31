# Frequency to Note

The closest note to a frequency, and how far off it is in cents.

## When to use it

- Finding the pitch of a sample, a resonance or a ringing drum.
- Checking how far a recorded instrument is from standard pitch.
- Naming a frequency you found with a spectrum analyser.

## How it works

The calculator compares the frequency with the notes of standard tuning (A4 = 440 Hz, twelve-tone equal temperament) and gives the difference in **cents**: 100 cents are one semitone. It shows the result in two ways:

- **Note / Cents** take the note at or below the frequency, and how far above it the frequency is: 0 to 99 cents.
- **Note (50¢n) / Cents (50¢n)** take the closest note, and how far the frequency is from it: −50 to +50 cents.

## How to use it

1. **Enter the frequency** in Hz, e.g. 440, 261.63 or 1000.
2. Optionally switch the **Middle C** convention (C3 or C4); it changes the octave numbers.
3. Or use the **Quickselect** buttons **C3**, **A3**, **C4** and **A4** to fill in the frequency of that note.
4. **Play** plays a 3-second test tone at the frequency on a computer; on Android the button does nothing. **Reset** returns to 440 Hz and the C3 convention.

## Reading the results

- **Note / Cents:** the note at or below the frequency, plus 0 to 99 cents.
- **Note (50¢n) / Cents (50¢n):** the closest note. Positive cents are sharp (higher), negative cents flat (lower). This is usually the one to tune by.

## In practice

- **Tuning a sample:** a sample at 367 Hz is F#3 −14 cents in 50-cent notation (C3 convention). To bring it to F#, raise it by 14 cents in the sampler.
- **Resonances:** name a room resonance or a ringing tom, to see whether it sits on a note of the song or between notes.

## Examples

- 440 Hz is A3 (C3 convention), 0 cents in both notations.
- 367 Hz is **F3 +86 cents**, or **F#3 −14 cents** in 50-cent notation.

## Good to know

- The reference is fixed at A4 = 440 Hz (A3 in the C3 convention). For other concert pitches or tunings, use **Note to Frequency**.
