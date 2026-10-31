# MusiCalc

A cross-platform music calculator application built with Go and Fyne. Provides real-time calculators for tempo, pitch, sampling, and time-stretching operations commonly used in music production and audio engineering. It runs on Windows, Linux and Android.

This project is heavily inspired by [MusicMath](https://dev.laurentcolson.com/musicmath.html), an excellent tool currently exclusive to the Apple ecosystem. This version aims to bring that same utility to Windows, Linux and Android, with a simpler look and feel. For macOS users looking for a polished commercial solution, MusicMath is highly recommended. MusicMath has long been a reference tool for music-related calculations, and this project draws inspiration from that work.

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

## Features

### ⏱️ Timecode Calculator
- Support for multiple frame rates: 23.976, 24, 25, 29.94, 29.97, 29.97 drop frame, 30, 50, 59.94, 60 fps
- Precise NTSC frame rate handling with exact fractional values
- Dual timecode input fields with real-time calculation
- Add and subtract timecode operations
- Frame count preservation when switching frame rates (H:M:S:F notation stays constant)
- Compact display format showing timecode, frame count, and FPS
- Calculation history of the last 100 operations (selectable and copyable on desktop)
- Essential for video editing, post-production, and audio-for-video work

### 🎵 Tempo to Delay Calculator
- Calculate delay times (ms) and modulation frequencies (Hz) for various note divisions
- One card per note value, from whole notes to 1/64 notes, with the straight, dotted and triplet values side by side
- Real-time BPM input with instant recalculation
- Built-in tap pad: tap along with the beat and the tempo field and all cards follow live (desktop: press T)
- Perfect for setting up delay effects, LFOs, and rhythmic modulation

### 👆 Tap Tempo
- Find the tempo of a song by tapping along with the beat on a large tap area
- Averages the last 8 taps; pausing for 2 seconds starts a new measurement
- Shows the tempo, the number of taps, how steady the taps were (± BPM) and the quarter-note length
- Round switches between two decimals and whole BPM
- Delay sends the measured tempo straight to the Tempo to Delay calculator
- Desktop: click, or press Space or T to tap; Esc resets

### 🎹 Note to Frequency Calculator
- Complete MIDI note range: C-2 to B8 (MIDI 0-131)
- Custom reference tuning support (default A3 = 440 Hz)
- 46 bundled tunings (Scala scales), organized into Equal, Just, Historic and World submenus
- Displays frequencies and cent deviations for all chromatic notes
- Highlights the reference note and opens scrolled to it
- Dual MIDI convention display (C4=60 standard / C3=60 alternative)
- Real-time frequency calculation with adjustable reference pitch
- Ideal for tuning synthesizers, creating custom scales, and frequency analysis

### 🎼 Frequency to Note Calculator
- Convert any frequency to the closest musical note
- Dual notation display: 100-cent and 50-cent systems
- Shows pitch deviation in cents for fine-tuning accuracy
- Quick-select buttons for common reference frequencies
- Adjustable middle C convention (C3/C4)
- Custom reference tuning support (default A4 = 440 Hz)
- Perfect for analyzing recordings, tuning acoustic instruments, and spectrum analysis

### ⏱️ Sample Length Calculator
- Bidirectional calculation: change any field and others update automatically
- Calculate sample count from tempo, beats, and sample rate
- Calculate tempo from sample length and beats
- Calculate duration (ms) from any combination of parameters
- Support for multiple sample rates (8kHz to 192kHz)
- Adjustable beat divisions for loop calculations
- Essential for sampler programming and loop creation

### 🎙️ Multi-Mic Alignment Delay
- Calculate delay offsets to time-align multiple microphones to a single reference mic
- Names the sound source all distances are measured from (e.g. Snare), so every number refers to the same thing
- Setup card with labelled source, temperature (°C/°F), reference mic distance and sample rate, pre-filled with useful defaults (Snare, 20 °C, 3 m, 48 kHz); it folds into a one-line summary such as "Snare · ref 3 m · 20 °C · 48 kHz" to leave room for the mics
- One distance unit (m/ft) for the reference, the microphones and the results
- Switching a unit converts the values (4 m → 13.12 ft, 22 °C → 71.6 °F), so the setup and all delays stay the same
- Per-mic delay to add, in milliseconds and samples (based on project sample rate)
- One compact card per microphone (name, distance, delay, samples); all delays update automatically when any input changes
- Mic names can be typed freely or picked from grouped presets (Drums, Overheads & Room, Instruments & Vocals)
- The mic at the reference distance is marked as the reference; a mic farther than the reference shows how much earlier it would have to be moved (e.g. −1.46 ms, 70 samples earlier)
- Warns when a stereo pair (e.g. Overhead L / Overhead R) gets different delays, which would shift the stereo image
- The last setup is restored when the app starts
- Presets for standard studio setups in the ⋮ menu: save, save as, load, rename and delete; import and export as a JSON file through the platform's own file dialog (Windows, macOS, Linux desktop portal, Android and iOS system picker)

### 〰️ Phase-Safe Distance
- Find the "sweet spots" and "death zones" for two mics on one source, e.g. a kick drum's 60 Hz fundamental
- Frequency input with grouped source presets (Drums, Bass, Guitars, Vocals) and air temperature (°C/°F)
- Shows the wavelength at that frequency and the speed of sound used
- Checks a path difference between the two mics: phase offset, zone (sweet spot, partial cancellation, death zone) and how much the frequency drops when both mics are summed
- Colour-coded zone map over two wavelengths, with the checked distance marked
- Lists the in-phase and opposed distances and how far you can stray from a sweet spot
- Distances in m or ft; switching converts the values
- Source, frequency, temperature and units sit in a Setup card that folds into one line (e.g. "Drums · Kick 60 Hz · 22 °C · m"); the setup is kept between sessions

### 📏 3:1 Rule
- Checks one source reaching two mics: its own mic and a neighbouring mic
- Setup card with the source, both distances, how much louder the neighbouring channel is (gain, sensitivity, pickup), the target (3:1 / 9.5 dB, 10 dB, 15 dB) and units; it folds into one line such as "Singer A 1 / 3 m · +6 dB · ≥ 10 dB"
- Separate results: the distance check (meets 3:1, at the threshold, below) and the estimated spill of the source in the neighbouring mic, against the target
- Concrete fixes: how many dB to turn the neighbouring channel down, or how far to move the neighbouring mic; or how much room there is to spare
- Comb-filter range stated as a condition (otherwise identical signals), including when full cancellation is possible
- Distances in m or ft; switching converts the values; the setup is kept between sessions

### 🎚️ Tempo Change Calculator
- Calculate pitch changes from tempo adjustments
- Time stretching percentage calculations
- Transpose semitones and cents (standard 100-cent notation)
- 50-cent notation display for granular pitch control
- Tempo delta percentage calculation
- Dynamic clamping to valid tempo range (5-1000 BPM)
- Bidirectional: adjust tempo, time stretch, or transpose values
- Swap tempo values with one click
- Perfect for time-stretching audio, pitch correction, and sampler tuning

## Documentation

- **[User Guide](README-USERGUIDE.md)** - How to use each calculator
- **[Ideas](IDEAS.md)** - Calculators that might be added in the future
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

MIT License - Copyright (c) 2026 B. Zeiss

See [LICENSE](LICENSE) for details.
