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
- Temperature input (C/F) to estimate speed of sound
- Distance units (m/ft) for reference and target microphones
- Per-mic results in milliseconds and samples (based on project sample rate)
- One card per microphone; all delays update automatically when any input changes
- Mic names can be typed freely or picked from grouped presets (Drums, Overheads & Room, Instruments & Vocals)
- If a target mic is further than the reference, the delay is shown as N/A

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
