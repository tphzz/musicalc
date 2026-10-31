# Ideas

Calculators and features that might be added to MusiCalc in the future.

Suggested order: quick access and shared session context, then Recording Time ↔ Storage, then Hardware Level & Calibration or Gain & Peak Headroom.

## New Calculators

- **Recording Time ↔ Storage:** Will a recording fit? From channel count, sample rate, bit depth (including 32-bit float) and duration, calculate the storage needed; from the free space, calculate the remaining recording time.
  - Count mono channels separately from stereo tracks.
  - Offer an optional storage reserve, and keep backup copies separate from the recording itself.
  - Examples: 32 mono channels · 48 kHz · 24-bit · 2 hours → 33.18 GB; at 96 kHz → 66.36 GB.
  - Reference for the practical scope: the Sound Devices recording-time calculator.
- **Hardware Level & Calibration:** Convert between dBu, dBV and volts RMS, and to dBFS using a converter calibration.
  - Saved, named input/output configurations, so the interface setting is picked rather than re-entered. Converters have selectable full-scale levels (e.g. RME).
  - Example: with 0 dBFS = +24 dBu, sending +4 dBu needs a −20 dBFS sine test tone, about 1.23 V RMS.
- **Gain & Peak Headroom:** Can a level change be done with gain alone? From the measured integrated loudness and true peak, the desired loudness and the peak ceiling, show:
  - the requested gain;
  - the predicted true peak at that gain;
  - the gain available before the ceiling;
  - the loudness at that gain.

  Labelled as a gain-only estimate based on the engineer's own measurements; any dynamics processing needs a new measurement.
  - Example: −18 LUFS / −3 dBTP, target −14 LUFS, ceiling −1 dBTP → requested +4 dB, predicted +1 dBTP, available +2 dB, −16 LUFS at that gain.
- **Acoustic Room Mode Calculator:** Predicts the resonant frequencies (standing waves) of a rectangular room from its length, width and height, with axial, tangential and oblique modes. Helps anticipate bass build-up and nulls, place acoustic treatment and find the listening position.
- **Air Absorption & Humidity Compensator:** Estimates the high-frequency loss of sound over distance from temperature and relative humidity, and the boost needed at given frequencies (e.g. 10 kHz) to restore distant mics' clarity, for large sessions and live sound.

## Extensions to Existing Calculators

- **Sample Length – sample/time offsets:** Direct conversion between samples and milliseconds for delay and track-offset fields, without tempo or loop settings. Example: 317 samples at 48 kHz → 6.604 ms.
- **Sample Length – bars ↔ duration ↔ tempo:** How long is a section, or which tempo fits a cue? The time signature and the BPM beat unit are explicit. Example: 32 bars of 4/4 at 120 BPM → 64 s.

## App-Wide Features

- **Shared session context:** One sample rate, tempo, frame rate and tuning used by all calculators.
- **Favorites and recent tools:** Each engineer puts their two or three regular calculations first (e.g. tracking and mixing engineers differ).
- **Immediate input:** Opening a calculator focuses the relevant field and restores its last state.
- **Tap a result to copy:** The number by default, with the unit available separately.
- **Fast desktop access:** A keyboard shortcut and an optional small floating window.
- **Offline operation:** Every core calculation available without a connection.

## Not Planned

- **Generic buffer-latency calculator:** Sample Length already converts samples to time; real monitoring latency also includes converter and system delays.
- **Automatic "ideal" compressor timings or reverb recipes:** These would claim more certainty than the maths supports, as the old "phase-safe" labels did.
