# MusiCalc User Guide

This guide explains each calculator, in the order the app shows them.

- [Getting Around](#getting-around)
- **Time & Tempo**: [Timecode](#timecode) · [Tempo to Delay](#tempo-to-delay) · [Tempo Change](#tempo-change) · [Tap Tempo](#tap-tempo)
- **Frequency & Pitch**: [Note to Frequency](#note-to-frequency) · [Frequency to Note](#frequency-to-note)
- **Analysis**: [Sample Length](#sample-length)
- **Multi-Mic**: [Alignment Delay](#alignment-delay) · [Two-Mic Sum](#two-mic-sum) · [3:1 Rule](#31-rule)

## Getting Around

- The **menu button** (☰) at the top left switches between the categories: **Time & Tempo**, **Frequency & Pitch**, **Analysis** and **Multi-Mic**.
- The **tab bar** at the bottom switches between the calculators of the current category. On Android the tabs show icons only.
- The header shows where you are, e.g. `Time & Tempo / Tap Tempo`. Calculators with extra actions, like the Alignment Delay presets, show a **⋮** button on the right of the header.
- Each calculator keeps its values when you switch to another one and back.
- On Android the tab bar hides while the on-screen keyboard is open and comes back when it closes.

**Setup cards (Multi-Mic).** The Multi-Mic calculators keep their settings in a **Setup card** at the top. Tap its header to fold it into one line that shows all values, e.g. **Singer A 1 / 3 m · +6 dB · ≥ 10 dB**, and tap again to open it. The setup and whether the card is open are kept when you close the app.

**Units.** Distances can be entered in metres or feet, temperatures in °C or °F. Switching a unit converts the values instead of re-interpreting them: 4 m becomes 13.12 ft, 22 °C becomes 71.6 °F, and every result stays the same.

---

# Time & Tempo

## Timecode

Add and subtract timecodes, and see their frame counts at different frame rates.

1. **Select the frame rate**: 23.976, 24, 25, 29.94, 29.97, 29.97 drop frame (df), 30, 50, 59.94 or 60 fps.
2. **Enter timecodes** in Hours:Minutes:Seconds:Frames:
   - Digits fill from right to left (frames → seconds → minutes → hours).
   - `.` (period) inserts `00`, e.g. `1.` = 1 second.
   - **Backspace** removes the last digit. Empty fields count as 0.
3. **Calculate**:
   - **+** adds Timecode 2 to Timecode 1, **-** subtracts it; the result moves to Timecode 1 for the next step.
   - **Reset** clears both timecodes and the history; **Clear Hist.** clears only the history.
4. **Switch the frame rate** to convert: the H:M:S:F notation stays, and the frame count is recalculated for the new rate. The conversion is logged in the history.
5. **Use the history**: every calculation and conversion is listed with its frame count and frame rate, newest first; the last 100 are kept. On desktop you can select and copy the text.

**Reading the display**: `00:34:35:04 (62254 frames @ 30)` is the timecode, its total frame count at the current rate, and the frame rate.

**Examples**:
- Add two clips: `00:22:22:22 + 00:12:12:12 = 00:34:35:04` at 30 fps.
- Convert: `00:34:35:04` is 62254 frames at 30 fps and 51879 frames at 25 fps.

## Tempo to Delay

Delay times and modulation rates for every note value at a tempo.

1. **Set the tempo** in the BPM field, or **tap it in** on the pad below the field:
   - Tap along with the beat (desktop: click the pad, or press **T** while the field isn't focused).
   - Each tap writes the measured tempo into the field, so all values update live.
   - Pausing for 2 seconds ends the measurement; the next tap starts a new one.
2. **Read the cards**: one card per note value, from 1/1 (whole note) to 1/64.
   - Each card lists the **Straight**, **Dotted** and **Triplet** delay in ms.
   - The value on the right is the matching modulation rate in Hz.
3. **Use the values** for delay times, LFO rates, gates and sequencers.

**Example**: at 120 BPM a quarter note (1/4) is 500 ms, or 2 Hz.

## Tempo Change

How a tempo change affects pitch and length, and the other way round. Enter any one of the values; the others follow.

1. **Set the Original Tempo** of your audio material.
2. **Enter one of**:
   - **New Tempo**: the tempo you want.
   - **Time stretching %**: the new length relative to the original. 100 % keeps the length; 200 % is twice as long (half the tempo); 50 % is half as long (double the tempo).
   - **Transpose Semis** and **Transpose Cents**: the pitch shift you want, as when a sample is played faster or slower without time-stretching.
3. **Read the results**:
   - **Transpose Semis/Cents**: the pitch shift in standard notation (100 cents = 1 semitone).
   - The 50-cent notation used by some samplers (50 cents = 1 semitone).
   - **Tempo Delta**: the tempo change in percent.
4. **Swap** exchanges the original and the new tempo; **Reset** returns to 120 → 100 BPM.

Results stay within 5–1000 BPM.

**Examples**:
- A 140 BPM loop played at 170 BPM: pitch +3.36 semitones (+3 semitones, +36 cents), time stretching 82.35 %.
- A sample pitched up 7 semitones plays about 1.5 times as fast: 120 BPM becomes 179.8 BPM, time stretching 66.74 %.

## Tap Tempo

Find the tempo of a song by tapping along with it.

1. **Tap the large area** along with the beat (desktop: click it, or press **Space** or **T**).
   - The tempo appears from the second tap and gets more precise with every tap; it is averaged over the last 8 taps.
   - The status line shows the number of taps, how steady they were (± BPM) and the length of a quarter note.
2. **Pause for 2 seconds** to end the measurement; the tempo stays on screen, and the next tap starts a new measurement.
3. **Use the buttons**:
   - **Round**: switches between two decimals and whole BPM (highlighted while rounding is on).
   - **Delay**: sends the tempo to Tempo to Delay and opens it.
   - **Reset**: clears the measurement (desktop: **Esc**).

**Tip**: tap at least 4 to 8 beats; a small ± value means your taps were steady and the tempo is reliable.

---

# Frequency & Pitch

## Note to Frequency

The frequency of every note, in standard and alternative tunings.

1. **Set the reference**: the reference note and its frequency (default A3 = 440 Hz with the C3 convention).
2. **Choose the middle C convention**: C3 or C4. It changes the octave numbers of the note names.
3. **Choose a tuning** (default Equal · 12-TET). The 46 tunings are grouped into submenus:
   - **Equal** (equal temperaments), **Just** (just intonation), **Historic** (well temperaments and meantones) and **World** (Arabic, Indian, Chinese, Japanese, Indonesian and other scales).
   - The selected tuning and its group are check-marked.
4. **Read the table**: every note from C-2 to B8 with its frequency, the deviation in cents from equal temperament, and the MIDI note number in both conventions (C3 = 60 and C4 = 60; MIDI covers 0–127).

The table opens scrolled to the reference note, which is highlighted. Every C is bold, marking the start of an octave. **Reset** returns to the defaults.

**Example**: to tune an 808 kick to your track's key, look up the root note's frequency and set the kick's pitch to it.

## Frequency to Note

The closest note to a frequency, and how far off it is. Notes are based on A4 = 440 Hz in equal temperament.

1. **Enter a frequency** in Hz, e.g. 440, 261.63 or 1000.
2. **Read the closest note** and its deviation in cents, in two notations:
   - **Note / Cents**: standard notation (100 cents = 1 semitone). Positive cents are sharp, negative cents flat.
   - **Note (50¢n) / Cents (50¢n)**: the notation some samplers use (50 cents = 1 semitone).
3. **Middle C**: switch between the C3 and C4 conventions for the octave numbers.
4. **Quickselect** (**C3**, **A3**, **C4**, **A4**) fills in the frequency of that note, following the middle C convention.
5. **Play** (desktop only) plays a 3-second test tone at the entered frequency. **Reset** returns to 440 Hz and the C3 convention.

**Example**: 367 Hz is F#3 −14 cents in standard notation, or F3 +86 cents in 50-cent notation.

---

# Analysis

## Sample Length

The length of a loop in samples and milliseconds, and its tempo. Change any field and the others follow.

1. **Set the Sample Rate**: pick one from 8000 to 192000 Hz, or type any rate.
2. **Set the Beats**: the loop length in beats (4 = one bar of 4/4, 16 = four bars).
3. **Enter one of**:
   - **Tempo**: to get the length in samples and ms.
   - **Length in samples**: to find the tempo that fits.
   - **Length in ms**: to get the tempo and the sample count.

**Reset** returns to 44100 Hz, 4 beats and 120 BPM.

**Examples**:
- An 88200-sample loop at 44.1 kHz is 2000 ms; as 4 beats it is 120 BPM.
- A four-bar loop (16 beats) at 140 BPM is 302400 samples (6857.14 ms) at 44.1 kHz.

---

# Multi-Mic

## Alignment Delay

Time-align close mics to a **reference** mic (usually the farthest one, e.g. a room mic): how much delay does each close mic need so that a sound arrives in all of them at the same time?

**All distances are measured from one sound source**, e.g. the snare. A room mic is at a different distance from the snare than from the kick, so one calculation always aligns the arrivals of one source. Aligning a mic for that source also shifts everything else it picks up.

1. **Check the Setup card**
   - **Source**: the sound source all distances are measured from. Type a name, or tap the arrow to pick one from **Drums** or **Instruments**.
   - **Temperature** (°C/°F): sets the speed of sound.
   - **Reference mic**: the distance of the reference mic from the source, in **m** or **ft**. This is the distance unit for the whole calculator.
   - **Sample rate**: your project's sample rate.
   - The first time, the card starts open with Snare, 20 °C (the usual reference temperature), 3 m and 48000 Hz. Once you add the first mic it folds into one line, e.g. **Setup Snare · ref 3 m · 20 °C · 48 kHz**.
2. **Add close mics**
   - **Name**: type one, or tap the arrow to pick a common position from **Drums**, **Overheads & Room** or **Instruments & Vocals**. Without a name a mic is called "Mic 1", "Mic 2", and so on.
   - **Distance** from the source; the field says what it is measured from, e.g. "From snare".
   - Press **+** to add the mic; the trash icon on its card removes it.
3. **Read the results**
   - The list is headed "Delay to add · estimate, check by ear": the values come from the distances, so listen to the result. Sample counts are rounded.
   - Each mic card shows the name and distance on the left, and the delay to add in **ms** and **samples** on the right.
   - The mic **at the reference distance** shows **0 ms · reference**.
   - A mic **farther than the reference** shows a negative value in amber, e.g. **−1.46 ms · 70 samples earlier**. A delay can't make it arrive earlier: move its track earlier by that amount in your DAW, or use that mic as the reference instead.
   - When a **stereo pair** (e.g. Overhead L and Overhead R) gets different delays, an amber note shows the difference. Delaying the two sides separately shifts the stereo image; to keep it, give both the same delay.

**Examples** (distances from the snare, 20 °C, reference 3 m, 48 kHz):
- Snare Top at 0.1 m needs 8.44 ms (405 samples); Overhead L at 1.2 m 5.24 ms (252 samples); Overhead R at 1.4 m 4.66 ms (224 samples). The overheads differ by 0.58 ms, so the pair note appears.
- Room L at 3 m is the reference (0 ms); Room R at 3.5 m arrives 1.46 ms (70 samples) later than the reference.

### Presets

For standard studio setups, store the setup as a preset. The **⋮** button at the top right opens the preset menu.

- A preset stores the **source, reference distance, units, sample rate and mics**. It does not store the temperature, which changes from day to day; loading a preset keeps the temperature you set.
- The setup card shows the loaded preset's name instead of "Setup", e.g. **Studio A kit**. An asterisk (**Studio A kit\***) means you changed something since loading it.
- **Save "Studio A kit"**: overwrites the loaded preset with your changes; only available when there are changes.
- **Save as…**: stores the setup under a new name. If the name already exists, you are asked whether to replace that preset.
- **Load preset…**: lists the presets with their source, reference distance and number of mics; tap one to load it. The loaded preset is checked.
- **Manage presets…**: rename (pencil) or delete (trash). A name that already exists is refused; deleting asks first and can't be undone.
- **Import presets…**: adds the presets from a MusiCalc preset file (.json). A preset whose name already exists is added as "Name (2)".
- **Export presets…**: saves all presets to one file, suggested as `musicalc-presets.json`, e.g. to copy them to another computer or phone.

Import and export use your system's own file dialog: the Windows Open/Save dialog, the macOS file panel, your Linux desktop's file chooser, or the Android/iOS file picker.

## Two-Mic Sum

What does mixing two mics on the same source do to the combined sound?

When two mics pick up the same source at different distances, the farther mic hears it later. Mixing them adds a delayed copy of the sound: some frequencies add up (**reinforcement**), others cancel (**cancellation**). This repeating pattern of peaks and dips is called comb filtering. The calculator shows what it does to the combined sound, and what a delay, level or polarity change would do, e.g. for a kick drum's close mic and a room mic.

The model is simple on purpose: the same signal reaching both mics, Mic B delayed and at the level and polarity you set. Room reflections and the mics' own response are not included, so always check both mics summed in mono.

1. **Check the Setup card**
   - **Source**: picks a typical **range of interest** and a frequency for the cursor, e.g. Kick 50–100 Hz with the cursor at 60 Hz, or Cymbals 3–12 kHz. Change the **Range** or **Frequency** to anything else and the source shows "Custom".
   - **Temperature** (°C/°F) sets the speed of sound; **Units** (m/ft) the distance unit; **Sample rate** the sample counts.
   - Folded, it reads e.g. **Drums · Kick 50–100 Hz · 20 °C · 48 kHz**.
2. **Describe Mic B compared with Mic A**
   - **Farther by**: how much farther Mic B is from the source than Mic A, as a distance or, if you measured it, as a time in **ms**. Switching between them converts the value through the speed of sound. A negative value means Mic B is closer.
   - **Level**: how loud Mic B is in the mix compared with Mic A, in dB (0 = equal, negative = quieter).
   - **Polarity**: Normal or Inverted (the polarity switch on the channel).
   - **Delay on Mic A**: a delay you apply to Mic A, in ms.
3. **Read the arrival difference**: e.g. **Mic B arrives 11.65 ms later**, 559 samples at 48 kHz. **Align** fills in that delay on Mic A to line up the direct sound; the card then says **Aligned**, and **Reset** removes the delay again.
4. **Read the graph**: the combined level over frequency, in dB relative to Mic A alone.
   - +6 dB means both mics add up fully; deep dips mean cancellation.
   - The range of interest is shaded and the cursor marks the selected frequency; the line below gives the level there and the lowest and highest level within the range.
   - The graph's frequency range follows the range of interest.
   - After you change the level, polarity or delay, a dashed line shows the response before the change.
   - Where the comb is too dense to draw as a line, the graph shows the band the level swings in.
5. **Read the predicted cancellation and reinforcement**: the first dip and peak frequencies from the graph's range on, how deep they go, and how far apart they repeat.
6. **Treat the suggestion as an option, not a rule**: it tells you which delay lines up the direct sound. Whether that sounds better is a listening decision; a room mic's later arrival can be part of the sound.

**Examples** (Mic B 4 m farther, 20 °C, 48 kHz, Kick 50–100 Hz):
- Equal level, normal polarity: Mic B arrives 11.65 ms (559 samples) later. At 60 Hz the sum is +1.4 dB; within 50–100 Hz it swings from −5.8 to +6.0 dB. Cancellation at 42.9 · 128.8 · 214.6 Hz, reinforcement at 85.9 · 171.7 · 257.6 Hz.
- **Align** (Mic A delayed by 11.65 ms): +6.0 dB at every frequency.
- Mic B at −6 dB with inverted polarity: dips only reach −6.0 dB and peaks +3.5 dB; inverting swaps where the dips and peaks are.

## 3:1 Rule

Does a source spill too strongly into a neighbouring mic?

The 3:1 rule is about **one source reaching two mics**: its own mic and a neighbouring mic (e.g. Singer A's mic and Singer B's mic). If the neighbouring mic is at least **three times as far** from the source as its own mic, the source's spill into the neighbouring mic is about 9.5 dB lower, with matched mics and gain. That keeps comb filtering shallow when both mics are mixed.

Check **one source at a time**, then do the same for each other source: one source passing doesn't mean the others do.

1. **Fill in the Setup card**
   - **Source**: the source you are checking, e.g. "Singer A" (optional; the results use the name).
   - **To its mic**: the distance from the source to its own mic.
   - **To neighbour mic**: the distance from the same source to the neighbouring mic.
   - **Neighbour gain**: how much louder the neighbouring channel is, in dB: more gain or fader, a more sensitive mic, or a pickup pattern that favours the source. Use a negative value if it is quieter. The default is 0 dB.
   - **Target**: how far below the spill should be: **3:1 (9.5 dB)**, **10 dB below** or **15 dB below**. These are criteria to aim for, not a promise that spill is inaudible; a full 10 dB needs 3.16 : 1.
   - **Units** (m/ft).
2. **Read the results**
   - The **distance ratio**, e.g. 3.0 : 1 (neighbour mic : its mic).
   - **Distance check**: ✓ **Meets the 3:1 guideline**, ! **At the 3:1 threshold** (exactly 3:1, no room to spare) or ✕ **Below 3:1**. This looks at the distances only.
   - **Estimated spill**: how far below (or above) its own mic the source reaches the neighbouring mic, including the neighbour gain, and whether that meets the target. It estimates the direct sound, assuming matched mics apart from the gain you entered.
   - **Room to spare**: how close the neighbouring mic may come while still meeting the target.
   - **To reach the target** (when it doesn't): turn the neighbouring channel down by the missing dB (this also lowers that channel's own source), move the neighbouring mic to at least the stated distance, or aim its rejection at the source. These are alternatives: try one, then check again.
   - **Comb filtering**: if both contributions were otherwise identical, how much mixing them could boost and cut at some frequencies. When the spill is as loud as its own mic or louder, some frequencies can cancel completely. Listen to both mics summed in mono.

**Examples** (Singer A):
- 1 m / 4 m, 0 dB, target 3:1: 4.0 : 1, meets the guideline; spill −12.0 dB, 2.5 dB better than 9.5 dB; the neighbouring mic may come as close as 3 m.
- 1 m / 3 m, neighbour 6 dB louder, target 10 dB: at the threshold, but the spill is only −3.5 dB, 6.5 dB short. Turn the neighbour down by 6.5 dB, or move it to at least 6.31 m.
- 1 m / 2 m, 0 dB, target 3:1: below 3:1, spill −6.0 dB, 3.5 dB short. Move the neighbouring mic to at least 3 m, or turn it down by 3.5 dB.
- 1 m / 2 m with the neighbour 6 dB quieter: still below 3:1, but the spill (−12.0 dB) meets the 9.5 dB target.
