# MusiCalc

A cross-platform music calculator built with Go and Fyne. It provides real-time calculators for timecode, tempo, pitch, sampling and multi-mic recording, the everyday maths of music production and audio engineering. It runs on Windows, Linux and Android.

This project is heavily inspired by [MusicMath](https://dev.laurentcolson.com/musicmath.html), an excellent tool currently exclusive to the Apple ecosystem. This version aims to bring that same utility to Windows, Linux and Android, with a simpler look and feel. For macOS users looking for a polished commercial solution, MusicMath is highly recommended. MusicMath has long been a reference tool for music-related calculations, and this project draws inspiration from that work.

## Screenshots

<table style="width: 100%;" border="0">
  <tr>
    <td>
      <img width="450" height="682" alt="image" src="https://github.com/user-attachments/assets/8b6ecfaf-d865-4699-9926-d56f9806eb15" />
    </td>
    <td>
      <img width="450" height="679" alt="image" src="https://github.com/user-attachments/assets/0f95847c-aa18-463b-a47b-81f80bf3f957" />
    </td>
    <td>
      <img width="450" height="679" alt="image" src="https://github.com/user-attachments/assets/2db75878-294e-425c-9ef3-1086adba30a7" />
    </td>
  </tr>
</table>

## Disclaimer

This application is provided for informational purposes only and is not guaranteed to be accurate. Calculations may contain errors due to software bugs, rounding, or floating-point limitations. Use this software entirely at your own risk; the author assumes no responsibility for any financial loss, data loss, or damages resulting from its use. Always manually verify critical results.

## Calculators at a Glance

| Category | Calculator | What it answers |
|---|---|---|
| Time & Tempo | ⏱️ Timecode | Add and subtract timecodes; convert between frame rates |
| | 🎵 Tempo to Delay | Delay times and modulation rates for every note value at a tempo |
| | 🎚️ Tempo Change | How a tempo change affects pitch and length, and the other way round |
| | 👆 Tap Tempo | The tempo of a song, tapped along with the beat |
| Frequency & Pitch | 🎹 Note to Frequency | The frequency of every note, in 46 tunings |
| | 🎼 Frequency to Note | The closest note to a frequency, and how far off it is |
| Analysis | 📐 Sample Length | Samples, milliseconds and tempo of a loop, from any one of them |
| Multi-Mic | 🎙️ Alignment Delay | The delay that lines each close mic up with a reference mic |
| | 〰️ Two-Mic Sum | What mixing two mics on one source does to the combined sound |
| | 📏 3:1 Rule | Whether a source spills too strongly into a neighbouring mic |

## Features

### Time & Tempo

#### ⏱️ Timecode
- Frame rates 23.976, 24, 25, 29.94, 29.97, 29.97 drop frame, 30, 50, 59.94 and 60 fps
- Two timecode fields with right-to-left digit entry and a `.` shorthand for `00`
- Add and subtract; the result moves to the first field for the next step
- Switching the frame rate keeps the H:M:S:F notation and recalculates the frame count
- History of the last 100 operations (selectable and copyable on desktop)

#### 🎵 Tempo to Delay
- Delay times (ms) and modulation rates (Hz) for note values from 1/1 to 1/64
- One card per note value with straight, dotted and triplet values side by side
- Built-in tap pad: tap along with the beat and every value follows live (desktop: press **T**)

#### 🎚️ Tempo Change
- Works in every direction: enter a new tempo, a time-stretch percentage, or a transposition in semitones and cents
- Shows the pitch shift in 100-cent and 50-cent notation, the time stretch and the tempo change in percent
- Limits results to the valid tempo range (5–1000 BPM); swap and reset buttons

#### 👆 Tap Tempo
- Large tap area; the tempo is averaged over the last 8 taps, and a 2-second pause starts a new measurement
- Shows the number of taps, how steady they were (± BPM) and the quarter-note length
- **Round** switches between two decimals and whole BPM; **Delay** sends the tempo to Tempo to Delay
- Desktop: click, or press **Space** or **T**; **Esc** resets

### Frequency & Pitch

#### 🎹 Note to Frequency
- Frequencies and cent deviations for every note from C-2 to B8, with MIDI numbers in both conventions (C3 = 60 and C4 = 60)
- Adjustable reference note and pitch (default A3 = 440 Hz in the C3 convention)
- 46 bundled tunings (Scala scales), grouped into Equal, Just, Historic and World submenus
- Opens scrolled to the highlighted reference note

#### 🎼 Frequency to Note
- The closest note to any frequency and its deviation in cents, in 100-cent and 50-cent notation (reference A4 = 440 Hz)
- Middle C convention C3 or C4; quick-select buttons for C3, A3, C4 and A4
- Desktop: plays a 3-second test tone at the entered frequency

### Analysis

#### 📐 Sample Length
- Change any field and the others follow: tempo, length in samples, length in ms
- Sample rates from 8 kHz to 192 kHz, or any rate you type; adjustable number of beats

### Multi-Mic

All three Multi-Mic calculators keep their settings in a **Setup card** that folds into a one-line summary, so the results get the screen, and restore the last setup when the app starts. Distances can be entered in metres or feet and temperatures in °C or °F; switching converts the values, so the physical setup stays the same.

#### 🎙️ Alignment Delay
- The delay each close mic needs to line up with the reference mic (usually a room mic), in ms and samples
- All distances are measured from one named source (e.g. the snare), so every number refers to the same thing
- One compact card per mic; mic names typed freely or picked from grouped positions
- The mic at the reference distance is marked; a mic farther away shows how much earlier it would have to move (e.g. −1.46 ms, 70 samples earlier)
- Warns when a stereo pair (e.g. Overhead L / R) gets different delays, which would shift the stereo image
- **Presets** for standard studio setups (⋮ menu): save, save as, load, rename, delete, and import/export as a JSON file through the system's own file dialog (Windows, macOS, Linux desktop portal, Android and iOS file picker)

#### 〰️ Two-Mic Sum
- Main result: how much later Mic B arrives than Mic A, in ms and samples, from the extra distance or a measured time
- Mic B's level and polarity, and a delay on Mic A; **Align** previews the delay that lines up the direct sound
- Frequency-response graph of the combined level, in dB relative to Mic A alone, with the range of interest shaded and a cursor; the graph follows the range of interest, and after a change shows the plain starting point (equal levels, no delay) as a dashed line
- Predicted cancellation (dips) and reinforcement (peaks) frequencies and their depth
- Source presets pick a typical range of interest (e.g. Kick 50–100 Hz, Cymbals 3–12 kHz)
- States its model and assumptions: the same signal reaching both mics, without room reflections; check the mics summed in mono

#### 📏 3:1 Rule
- Checks one source reaching its own mic and a neighbouring mic
- Separate results: the distance check (meets 3:1, at the threshold, below) and the estimated spill into the neighbouring mic, including how much louder that channel is
- Targets 3:1 (9.5 dB), 10 dB or 15 dB below
- Concrete fixes: how many dB to turn the neighbouring channel down or how far to move the mic; or how much room there is to spare
- Comb-filter range stated as a condition, including when full cancellation is possible

### Across the App
- Categories in the ☰ menu, calculators in the tab bar; each calculator keeps its values when you switch away and back
- Built-in help for every calculator in its ⋮ menu (desktop: **F1**, **Esc** closes): what it does, how it works, how to use it in a session, with examples and illustrations; the same pages are linked under [Help](#help) below
- Phone layout on Android: icon-only tabs, numeric keyboards, and a tab bar that hides while the keyboard is open
- Light and dark theme following the system

## Help

The same pages as the app's built-in help: what each calculator does, how it works, and how to use it in practice.

- [Getting Around](internal/ui/help/getting-around.md): navigation, menus and keys
- **Time & Tempo:** [Timecode](internal/ui/help/timecode.md) · [Tempo to Delay](internal/ui/help/tempo-to-delay.md) · [Tempo Change](internal/ui/help/tempo-change.md) · [Tap Tempo](internal/ui/help/tap-tempo.md)
- **Frequency & Pitch:** [Note to Frequency](internal/ui/help/note-to-frequency.md) · [Frequency to Note](internal/ui/help/frequency-to-note.md)
- **Analysis:** [Sample Length](internal/ui/help/sample-length.md)
- **Multi-Mic:** [Alignment Delay](internal/ui/help/alignment-delay.md) · [Two-Mic Sum](internal/ui/help/two-mic-sum.md) · [3:1 Rule](internal/ui/help/three-to-one.md)

## Documentation

- **[Build Instructions](build/docs/README-BUILD.md)** - How to build from source on Windows, Linux and for Android
- **[Creating Installers](build/docs/README-INSTALLER.md)** - Creating Windows installers and Linux packages
- **[Icon Resources](build/docs/README-ICONS.md)** - Icon specifications and resources
- **[Test Instructions](build/docs/README-TESTS.md)** - How to run tests

## Requirements

- Go 1.26.2 or later
- GCC/MinGW (for CGO support on Windows)
- Fyne v2.8.2
- For Android: Android SDK with NDK and the Fyne CLI (see the build instructions)

## License

MIT License - See [LICENSE](LICENSE) for details.
