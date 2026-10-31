# Tap Tempo

Find the tempo of a song by tapping along with it.

## When to use it

- Finding the tempo of a reference track, a live band or a sample.
- Checking whether a recording drifts in tempo.
- Getting a tempo quickly to set delays (see **Tempo to Delay**).

## How it works

Each tap marks a beat. The time between two taps gives a tempo: 60000 ÷ milliseconds between taps. The calculator averages the last 8 intervals, so the tempo gets steadier the longer you tap. Taps closer together than 150 ms count as one (an accidental double tap), and a pause of 2 seconds ends the measurement.

## How to use it

1. **Tap the large area** along with the beat. On a computer you can also click it, or press **Space** or **T**.
2. **Keep tapping** for at least 4 to 8 beats; the tempo appears from the second tap.
3. **Pause for 2 seconds** to end the measurement. The tempo stays on screen; the next tap starts a new measurement.

## Reading the results

- The large number is the **tempo** in BPM.
- The line below shows the **number of taps**, how **steady** they were (± BPM), and the length of a **quarter note**.
- A small ± value means your taps were even and the tempo is reliable.

## In practice

- Tap on the **quarter notes** (the beat you would count), not on every eighth note.
- **Round** switches between two decimals and whole BPM. Most songs have a whole-number tempo; rounding gives you that number directly.
- **Delay** sends the tempo to **Tempo to Delay** and opens it.
- **Reset** clears the measurement (on a computer also **Esc**).

## Examples

- Taps every 500 ms give **120 BPM**, and a quarter note of 500 ms.
- Taps every 600 ms give **100 BPM**.

## Good to know

- If the result is half or double what you expect, you tapped every other beat or on the eighth notes; halve or double it.
- Songs played live often drift by a few BPM; the ± value shows how much.
