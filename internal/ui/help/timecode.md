# Timecode

Add and subtract timecodes, and see how many frames they are at any frame rate.

## When to use it

- Adding up the lengths of clips, cues or reels.
- Finding the distance between two timecodes, e.g. a cue's start and end.
- Turning a timecode into a frame count, for tools that work in frames.
- Checking how many frames a timecode is at another frame rate.

## How it works

A timecode `HH:MM:SS:FF` counts hours, minutes, seconds and frames. How many frames that is depends on the frame rate: one second is 25 frames at 25 fps and 30 frames at 30 fps. The calculator turns both timecodes into frame counts, adds or subtracts them, and turns the result back into a timecode.

**Drop frame** (29.97 df) keeps NTSC timecode in step with the clock. The video runs at 29.97 frames per second, slightly slower than 30, so frame numbers 00 and 01 are skipped at the start of every minute, except every tenth minute. The calculator takes these skipped numbers into account.

## How to use it

1. **Pick the frame rate**: 23.976, 24, 25, 29.94, 29.97, 29.97 drop frame (df), 30, 50, 59.94 or 60 fps.
2. **Type the first timecode.** Digits fill from the right, like a calculator: frames first, then seconds, minutes and hours.
3. **Type the second timecode** the same way.
4. **Press +** to add the second timecode to the first, or **-** to subtract it. The result moves into the first field, ready for the next step.

| Key | What it does |
|---|---|
| Digits | Fill from right to left: frames, seconds, minutes, hours |
| **.** (period) | Inserts `00`, e.g. `1.` is 1 second |
| Backspace | Removes the last digit |

An empty field counts as 0.

## Reading the results

- Next to each timecode you see its **frame count** at the current frame rate, e.g. `(62254f @ 30)`.
- The **history** below lists every calculation and frame-rate change, newest first, with frame counts. It keeps the last 100 entries; on a computer you can select and copy them.
- **Switching the frame rate** keeps the written timecode and recalculates its frame count, and logs the conversion in the history.

## In practice

- **Adding several clips:** type the first clip, then the second, press **+**; type the third, press **+** again, and so on. The running total stays in the first field.
- **Length of a cue:** type the end time first and the start time second, then press **-**.
- **Frame counts for other tools:** set the project's frame rate first, then read the frame count next to the timecode.
- **Reset** clears both timecodes and the history; **Clear Hist.** clears only the history.

## Examples

- Two clips at 30 fps: `00:22:22:22 + 00:12:12:12 = 00:34:35:04`, which is 62254 frames.
- The same `00:34:35:04` at 25 fps is 51879 frames.
- At 29.97 df, `00:10:00:00` is 17982 frames: 18 frame numbers were skipped in those ten minutes.

## Good to know

- Timecode numbers whole frames, so 23.976, 29.97 and 59.94 fps count frames exactly like 24, 30 and 60 fps. Only **29.97 df** counts differently, because it skips frame numbers. **29.94** counts the same as 29.97.
- The calculator works with timecode and frame counts, not with clock time: it doesn't tell you how many real seconds a timecode lasts at an NTSC rate.
- Changing the frame rate keeps the timecode as written and only counts its frames again; it doesn't convert a position or a duration from one rate to another.
