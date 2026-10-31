# User Guide

## Getting Around

- The **menu button** (☰) at the top left switches between the categories: **Time & Tempo**, **Frequency & Pitch**, **Analysis** and **Multi-Mic**.
- The **tab bar** at the bottom switches between the calculators of the current category. On Android the tabs show icons only.
  - **Time & Tempo**: Timecode, Tempo to Delay, Tempo Change, Tap Tempo
  - **Frequency & Pitch**: Note to Frequency, Frequency to Note
  - **Analysis**: Sample Length
  - **Multi-Mic**: Alignment Delay, Phase-Safe Distance, 3:1 Rule
- The header shows where you are, e.g. `Time & Tempo / Tap Tempo`.
- On Android the tab bar hides while the on-screen keyboard is open and comes back when it closes.
- Each calculator keeps its values when you switch to another one and back.

## Timecode

1. **Select frame rate** from the dropdown (23.976, 24, 25, 29.97, 30, etc.)
2. **Enter timecode values** in Hours:Minutes:Seconds:Frames format
   - Fields start empty - just click and type
   - **Right-justified entry**: Type digits and they fill from right to left (frames → seconds → minutes → hours)
   - **Dot shorthand**: Press `.` (period) to quickly insert `00` - great for fast entry (e.g., `1.` = 1 second)
   - **Backspace** removes the last entered digit
   - Empty fields are treated as 0 in calculations

3. **Perform calculations**:
   - **Add**: Combines Timecode 1 and Timecode 2, moves result to Timecode 1
   - **Subtract**: Subtracts Timecode 2 from Timecode 1, moves result to Timecode 1
   - **Reset**: Clears all timecode fields and history
   - **Clear History**: Clears only the history, keeps current timecode values

4. **Switch frame rates**: Change the FPS dropdown to see:
   - Same timecode notation (H:M:S:F stays constant)
   - Recalculated frame count for the new frame rate
   - History entry showing the conversion

5. **View history**: All calculations and conversions are logged with frame counts and FPS
   - The newest entry is at the top; the last 100 operations are kept
   - On desktop, select history text to copy it for documentation
   - History persists until cleared or reset

**Understanding the Display**:
- Format: `00:34:35:04 (62254 frames @ 30)`
- First part: Hours:Minutes:Seconds:Frames
- Parentheses: Total frame count at current FPS
- After @: Current frame rate

**Example Use Cases**:
- Add two video clips: `00:22:22:22 + 00:12:12:12 = 00:34:35:04 @ 30 fps`
- Convert timecode between formats: `00:34:35:04 (62254f) @30 → 00:34:35:04 (51879f) @25`
- Calculate exact frame counts for EDL or AAF workflows
- Verify timecode calculations for post-production deliverables

## Tempo to Delay

1. **Set your project tempo** in the BPM field at the top, or **tap it in**:
   - Tap the pad below the field along with the beat (desktop: click it, or press **T** while the field isn't focused)
   - Each tap writes the measured tempo into the field, so all values update live
   - Pausing for 2 seconds ends the measurement; the next tap starts a new one
2. **View delay times**: there is one card per note value, from 1/1 (whole note) to 1/64
   - Each card lists the **Straight**, **Dotted** and **Triplet** delay in ms
   - The value on the right is the matching modulation rate in Hz
3. **Use the values** to configure:
   - Delay effect times for rhythmic echoes
   - LFO rates for synchronized modulation
   - Gate/sequencer timing

**Example**: At 120 BPM, a 1/4 note = 500 ms, perfect for a quarter-note delay

## Tap Tempo

Find the tempo of a song by tapping along with it.

1. **Tap the large area** along with the beat (desktop: click it, or press **Space** or **T**)
   - The tempo appears from the second tap and gets more precise with every tap
   - It is averaged over the last 8 taps
   - The status line shows the number of taps, how steady they were (± BPM) and the length of a quarter note
2. **Pause for 2 seconds** to end the measurement; the tempo stays on screen, and the next tap starts a new measurement
3. **Use the buttons**:
   - **Round**: switches between two decimals and whole BPM (the button is highlighted while rounding is on)
   - **Delay**: sends the tempo to the Tempo to Delay calculator and opens it
   - **Reset**: clears the measurement (desktop: **Esc**)

**Tip**: Tap at least 4 to 8 beats; a small ± value means your taps were steady and the tempo is reliable.

## Note to Frequency

1. **Set reference frequency** (default: 440 Hz)
2. **Set the reference note** if using non-standard tuning (default: A3)
3. **Choose middle C convention**: Switch between C3 and C4 naming conventions
4. **Choose a tuning** from the dropdown (default: Equal · 12-TET)
   - The 46 tunings are grouped into submenus: **Equal** (equal temperaments), **Just** (just intonation), **Historic** (well temperaments and meantones) and **World** (Arabic, Indian, Chinese, Japanese, Indonesian and other scales)
   - The selected tuning and its group are check-marked
5. **Browse the frequency table** to find:
   - Exact frequencies for synthesizer tuning
   - The deviation in cents from equal temperament
   - MIDI note numbers for programming (both conventions)
   - Pitch relationships between notes

The table opens scrolled to the reference note, which is highlighted. Every C is shown in bold to mark the start of an octave.

**Example**: To tune a 808 kick to your track's key, find the root note frequency and adjust the kick's pitch to match

## Frequency to Note

1. **Enter frequency** in Hz (e.g., 440, 261.63, 1000)
2. **View the closest note** with pitch deviation in cents, in both notation systems:
   - **Note / Cents**: Standard music notation (100 cents = 1 semitone)
   - **Note (50¢n) / Cents (50¢n)**: Alternative notation used by some samplers (50 cents = 1 semitone)

3. **Adjust settings** (optional):
   - **Middle C**: Switch between C3 and C4 conventions

4. **Use the Quickselect buttons** (**C3**, **A3**, **C4**, **A4**) to fill in the frequency of that note
   - The frequencies follow the selected Middle C convention
   - Instantly see the note and any pitch deviation

5. **Play** (desktop only) plays a short test tone at the entered frequency

**Understanding the Output**:
- **100 Cents Notation**: Shows closest note and deviation (e.g., "F#3, -14 cents")
  - Positive cents = sharp (higher than note)
  - Negative cents = flat (lower than note)
- **50 Cents Notation**: Alternative display (e.g., "F3, +86 cents")
  - Same pitch, different reference point

**Example Use Cases**:
- Analyze a synth tone at 367 Hz → Result: F#3, -14 cents (or F3, +86 cents)
- Find what note a resonant frequency represents
- Tune acoustic instruments by analyzing recorded frequencies
- Identify pitch of sampled sounds for proper key mapping
- Match synth oscillators to specific frequencies

## Sample Length

1. **Choose your workflow**:
   - Enter **Tempo** to calculate sample length for a specific BPM
   - Enter **Length in samples** to find what tempo matches your loop
   - Enter **Length in ms** to calculate both tempo and sample count

2. **Set parameters**:
   - **Sample Rate**: Select your project's sample rate (typically 44100 or 48000)
   - **Beats**: Set loop length (4 = 4-beat loop, 8 = 8-beat loop, etc.)

3. **Read calculated values** instantly in all other fields

**Example Use Cases**:
- You have a 88200-sample loop at 44.1kHz → Calculator shows it's 2000ms and fits 120 BPM at 4 beats
- You want a 4-bar loop at 140 BPM → Calculator shows you need 151,543 samples at 44.1kHz
- You need to know what tempo your found sample is → Enter its sample count and get the BPM

## Multi-Mic Alignment Delay

This calculator helps you time-align multiple microphones (close mics) to a single **reference** microphone (usually the furthest mic, e.g. a room mic). It computes how much delay to add to each close mic so that all transients arrive at the DAW at the same time.

1. **Check the Setup card**
   - The card at the top holds the settings, each with its label: **Temperature**, **Reference mic** and **Sample rate**.
   - It starts open with useful defaults: 20 °C (the usual reference temperature for the speed of sound), 3 m and 48000 Hz.
   - **Temperature**: the room temperature in **°C** or **°F**; it sets the speed of sound.
   - **Reference mic**: the distance of your farthest microphone from the source (the one you align everything to, usually a room mic), in **m** or **ft**. This is the distance unit for the whole calculator: the reference, new mics and the results.
   - **Sample rate**: your DAW/project sample rate (e.g. 44100, 48000, 96000).
   - Once you add the first mic, the card folds into one line, e.g. **Setup 20 °C · ref 3 m · 48 kHz**. Tap it to open it again.

2. **Add close microphones**
   - Enter a mic name: type a custom one, or tap the arrow in the Name field to pick a common position from the **Drums**, **Overheads & Room** or **Instruments & Vocals** submenu.
   - Enter its distance from the source in the calculator's unit (shown next to the field).
   - Press **+** to add it. Without a name it is called "Mic 1", "Mic 2", and so on.

3. **Read results**
   - Each mic gets a compact card: its name and distance on the left, the required delay in **milliseconds (ms)** and **samples** on the right.

4. **Remove microphones**
   - Tap the trash icon on a mic's card to remove it.

**Switching units** converts the values instead of re-interpreting them: with the reference at 4 m, switching to **ft** shows 13.12 ft, and a mic at 1 m shows 3.28 ft. Switching **°C** to **°F** turns 22 °C into 71.6 °F. Your setup, and therefore every delay, stays exactly the same.

**Notes / edge cases**:
- If a mic is **further than the reference**, the delay is shown as **N/A** because you cannot add a positive delay to make a further mic arrive earlier.
- All values update automatically when you change temperature, units, sample rate, reference distance, or add/remove microphones.

**Examples** (48 kHz):
- Defaults (20 °C, reference 3 m): Kick In at 1 m needs 5.82 ms (280 samples), Kick Out at 1.5 m 4.37 ms (210 samples).
- 22 °C, reference 4 m: Kick In at 1 m needs 8.7 ms (418 samples), Kick Out at 1.5 m 7.25 ms (348 samples).

## Phase-Safe Distance

When two mics pick up the same source at different distances, the farther mic hears it later. For a given frequency this delay is a phase offset: at some distances the two mics reinforce each other (**sweet spots**), at others that frequency cancels out (**death zones**). This calculator maps those distances, e.g. for a kick drum's 60 Hz fundamental.

1. **Set the frequency** in Hz, or pick a typical fundamental from **Source** (Drums, Bass, Guitars, Vocals). Typing another frequency shows "Custom".
2. **Set the temperature** (°C/°F) and the distance **Units** (m/ft). Switching converts the values.
3. **Read the wavelength**: e.g. 5.74 m at 60 Hz and 22 °C, with half a wavelength and the speed of sound used.
4. **Enter the path difference**: how much farther one mic is from the source than the other.
   - The card turns **green** (sweet spot, under 60°), **amber** (partial cancellation, 60–120°) or **red** (death zone, over 120°).
   - It shows the phase offset and how much the frequency drops when both mics are summed.
5. **Use the zone map**: it covers two wavelengths, shows green, amber and red ranges, and marks your path difference.
6. **Read the zone list**: the exact in-phase distances (0, λ, 2λ) and opposed distances (½ λ, 1½ λ), and how far you may be from a sweet spot (±⅙ λ, ±60°).

**Examples** (60 Hz, 22 °C, λ = 5.74 m):
- Path difference 0.8 m → 50°, sweet spot, −0.9 dB when summed.
- Path difference 2.6 m → 163°, death zone, −16.6 dB when summed: move a mic or flip its polarity.

## 3:1 Rule

The 3:1 rule says: another source should be at least **three times farther** from a mic than the mic's own source. Its bleed then arrives at least 9.5 dB quieter, which keeps comb filtering shallow when the mics are mixed.

1. **Enter the distance from the mic to its own source** and **to the other source**.
2. **Choose the Units** (m/ft). Switching converts both values.
3. **Read the result**:
   - The **distance ratio** (e.g. 4.0 : 1), green when the rule is met and red when it isn't.
   - How much quieter (or louder) the other source arrives at this mic, in dB.
   - **Margin**: how close the other source may come, and how far the mic may move from its own source, while still meeting the rule.
   - **To fix it** (when the rule is broken): move the other source to at least 3 × the own-source distance, or bring the mic closer to its own source.

**Examples**:
- Own source 0.3 m, other source 1.2 m → 4.0 : 1, bleed 12.0 dB below; the other source may come as close as 0.9 m.
- Own source 0.3 m, other source 0.6 m → 2.0 : 1, bleed only 6.0 dB below; move the other source to at least 0.9 m or the mic to within 0.2 m.

## Tempo Change

1. **Set original tempo** of your audio material
2. **Choose your operation**:
   
   **Option A - Change Tempo Directly**:
   - Enter **New Tempo** to see pitch shift required
   - View **Time Stretch %** (e.g., 200% = double speed)
   - See **Tempo Delta** (percentage change)
   
   **Option B - Change Time Stretch**:
   - Enter **Time Stretching %** to see resulting tempo
   - View pitch changes in semitones and cents
   
   **Option C - Transpose**:
   - Enter **Transpose Semis** and **Cents** for desired pitch shift
   - View resulting tempo and time stretch amount

3. **Use the Swap button** to quickly reverse tempo/new tempo values
4. **Use Reset** to return to default values (Original Tempo: 120, New Tempo: 100)

**Understanding the Output**:
- **Transpose Semis/Cents**: Standard pitch notation (100 cents = 1 semitone)
- **50 Cents Notation**: Alternative pitch display (50 cents = 1 semitone) used by some samplers
- **Tempo Delta**: Percentage change from original tempo (+/- %)

**Example Use Cases**:
- Speed up a 140 BPM loop to 170 BPM → See it shifts the pitch by about +3.36 semitones (+3 semitones, +36 cents)
- You need to pitch a sample up 7 semitones → See it will play at 150% speed
- Match a sample's tempo to your project without changing pitch → Use time-stretching calculations
