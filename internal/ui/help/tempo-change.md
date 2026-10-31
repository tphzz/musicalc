# Tempo Change

How a tempo change affects the pitch and length of audio, and the other way round: which tempo or stretch belongs to a pitch shift.

## When to use it

- Speeding up or slowing down a sample or loop to fit a project, the way a sampler or turntable does: tempo and pitch change together.
- Working out how far a sample is transposed when you play it at another tempo.
- Finding the time-stretch percentage a stretching tool needs.

## How it works

When audio is played faster without time-stretching, it gets higher; slower, it gets lower. Doubling the speed doubles the tempo and raises the pitch by an octave (12 semitones).

- **Pitch shift** in semitones = 12 × log2(new tempo ÷ original tempo).
- **Time stretching %** is the new **length** compared with the original: original tempo ÷ new tempo × 100. Above 100 % the audio gets longer and slower, below 100 % shorter and faster.
- **Tempo Delta** is the tempo change in percent.

## How to use it

1. **Set the Original Tempo** of your audio.
2. **Enter one of these**; the others follow:
   - **New Tempo**: the tempo you want.
   - **Time stretching %**: the length you want, relative to the original.
   - **Transpose Semis** and **Transpose Cents**: the pitch shift you want.
3. **Swap** exchanges the original and the new tempo. **Reset** returns to 120 → 100 BPM.

## Reading the results

- **Transpose Semis / Cents:** the pitch shift in standard notation (100 cents = 1 semitone).
- **50-cent notation:** the same shift as some samplers show it (50 cents = 1 semitone).
- **Time stretching %:** 100 % keeps the length; 200 % is twice as long (half the tempo); 50 % is half as long (double the tempo).
- **Tempo Delta:** how much faster (+) or slower (−) the new tempo is, in percent.

## In practice

- **Re-pitching a loop** to the project tempo: enter the loop's tempo and the project tempo, then set the sampler's transpose to the semitones and cents shown.
- **Small changes** (a few BPM) shift the pitch by less than a semitone: often fine for drums, audible on tonal material.
- **To keep the pitch** while changing the tempo, use a time-stretching tool with the **Time stretching %** shown, instead of re-pitching.

## Examples

- A 140 BPM loop played at 170 BPM: **+3.36 semitones** (+3 semitones, +36 cents), time stretching **82.35 %**.
- A sample pitched up **7 semitones** plays about 1.5 times as fast: 120 BPM becomes **179.8 BPM**, time stretching 66.74 %.

## Good to know

- Results stay within 5–1000 BPM.
- The calculator describes re-pitching (tempo and pitch together). Modern time-stretching changes the tempo without changing the pitch; then only **Time stretching %** applies.
