# Tap Tempo

Find the tempo of a song by tapping along with it.

## When to use it

- Finding the tempo of a reference track, a live band or a sample.
- Getting a tempo quickly to set delays (see **Tempo to Delay**).

## How it works

Each tap marks a beat. The time between two taps gives a tempo: 60000 ÷ milliseconds between taps. The calculator averages up to the last 8 intervals (9 taps): more taps even out small timing errors, and after that the tempo follows your most recent taps. Taps closer together than 150 ms count as one (an accidental double tap), and a pause of 2 seconds ends the measurement.

## How to use it

1. **Tap the large area** along with the beat. On a computer you can also click it, or press **Space** or **T**.
2. **Keep tapping** for at least 4 to 8 beats; the tempo appears from the second tap.
3. **Pause for 2 seconds** to end the measurement. The tempo stays on screen; the next tap starts a new measurement.

## Reading the results

- The large number is the **tempo** in BPM.
- The line below shows the **number of taps**, how **steady** they were (± BPM), and the length of a **quarter note**.
- **±** is how much the tempo of single intervals varies around the average. It measures your tapping as much as the music: a small value means your last taps were even, not that you tapped the right beat.

## In practice

- Tap on the **quarter notes** (the beat you would count), not on every eighth note.
- **Round** switches between two decimals and whole BPM. Songs produced to a click usually have a whole-number tempo, and rounding gives you that number; live recordings often don't.
- **Delay** sends the tempo to **Tempo to Delay** and opens it.
- **Reset** clears the measurement (on a computer also **Esc**).

## Examples

- Taps every 500 ms give **120 BPM**, and a quarter note of 500 ms.
- Taps every 600 ms give **100 BPM**.

## Good to know

- If the result is half or double what you expect, you tapped every other beat or on the eighth notes; halve or double it.
- Tapping gives the average tempo of your last taps. It can't follow tempo changes within a song, and it is only as accurate as your timing: at 120 BPM over 8 intervals, a last tap 10 ms late reads about 119.7 BPM.
