# Sample Length

The length of a loop in samples and milliseconds, and the tempo it fits. Change any field and the others follow.

## When to use it

- Cutting a loop to an exact number of beats in a sampler or editor.
- Finding the tempo of a loop from its length in samples.
- Checking how many samples a bar is at a tempo and sample rate.

## How it works

A beat lasts 60 ÷ BPM seconds. A loop of a number of beats therefore lasts beats × 60 ÷ BPM seconds, and that time multiplied by the sample rate is its length in samples. The calculator solves this in any direction: from the tempo, from the samples or from the milliseconds.

## How to use it

1. **Set the Sample Rate**: pick one from 8000 to 192000 Hz, or type any rate.
2. **Set the Beats**: the loop's length in beats (4 = one bar of 4/4, 16 = four bars).
3. **Enter one of**: the **Tempo**, the **Length in samples** or the **Length in ms**. The other two follow.

**Reset** returns to 44100 Hz, 4 beats and 120 BPM.

## Reading the results

- **Tempo:** the BPM at which the loop is exactly the given number of beats.
- **Length in samples:** the sample count, e.g. for a sampler's start and end points.
- **Length in ms:** the duration.

## In practice

- **Finding a loop's tempo:** read its length in samples from your editor, set the number of beats you hear, and read the tempo.
- **Cutting a loop:** set the tempo and beats, and cut at the sample count shown.
- Count beats carefully: if the tempo comes out half or double what you expect, the number of beats is off by a factor of two.

## Examples

- An 88200-sample loop at 44.1 kHz is 2000 ms; as 4 beats it is **120 BPM**.
- A four-bar loop (16 beats) at 140 BPM is **302400 samples** (6857.14 ms) at 44.1 kHz.

## Good to know

- A tempo rarely gives a whole number of samples; the sample count shows whole samples and drops any fraction.
