# Getting Around

MusiCalc groups its calculators into four categories. Each calculator answers one everyday question of music production and recording, and shows its results while you type.

## Finding a calculator

- The **menu button** (☰) at the top left switches between the categories: **Time & Tempo**, **Frequency & Pitch**, **Analysis** and **Multi-Mic**.
- The **tab bar** at the bottom switches between the calculators of the current category. On Android and iPhone the tabs show icons only.
- The header shows where you are, e.g. `Time & Tempo / Tap Tempo`.

| Category | Calculators |
|---|---|
| Time & Tempo | Timecode, Tempo to Delay, Tempo Change, Tap Tempo |
| Frequency & Pitch | Note to Frequency, Frequency to Note |
| Analysis | Sample Length |
| Multi-Mic | Alignment Delay, Two-Mic Sum, 3:1 Rule |

## Help and actions

- The **⋮** button at the top right opens the calculator's menu. Some calculators have their own actions there, like the Alignment Delay presets; **Help** is always the last entry.
- Help shows the page for the calculator you are using. The **✕** at the top right closes it again.
- On a computer, **F1** opens help and **Esc** closes it. On Android, the back button closes it.

## App version

Open **About MusiCalc** from the top-left menu. On macOS, use **MusiCalc > About MusiCalc** in the system menu bar instead. The dialog shows the version; packaged mobile builds also show the build number in parentheses. **Close**, **Esc**, or Android Back dismisses it without changing your calculator values.

## Things that work the same everywhere

- **Your values stay.** Each calculator keeps its values when you switch to another one and back, and when you open help. Only the Multi-Mic calculators also keep them when you close the app.
- **Setup cards.** The Multi-Mic calculators keep their settings in a **Setup** card at the top. Tap its header to fold it into one line that still shows every value, and tap again to open it. The setup is kept when you close the app. Two-Mic Sum and 3:1 Rule also remember whether the card was open; Alignment Delay opens it only while it has no mics.
- **Units convert.** Distances can be entered in metres or feet, temperatures in °C or °F. Switching a unit converts the shown values instead of re-interpreting them: 4 m becomes 13.12 ft, 22 °C becomes 71.6 °F, and every result stays the same.
- **On Android and iPhone** the tab bar hides while the on-screen keyboard is open and comes back when it closes.

## Desktop display scale

MusiCalc uses the desktop display scale for its physical size. On Windows, choose **Settings > System > Display > Scale** for the monitor in use. A wider resolution alone does not select a larger scale.

## Preferences

Open **Preferences…** in the top-left menu. A separator groups Preferences and About below the calculator categories.

**Language:** choose **System**, **English** or **Deutsch**. System is the default: German OS locales use German; other or unavailable locales use English. An explicit language choice overrides detection and stays saved. Changes apply immediately to the interface and help without resetting inputs, results, history or the selected calculator. Native OS controls retain the OS language.

**Interface scale (desktop):** choose **80%, 100% (default), 115%, 130% or 200%**. The percentage adjusts MusiCalc relative to the desktop's OS scale and is saved immediately. The whole interface updates without a restart. Fyne rounds the final display scale to 10% steps: for example, 115% on a 100% desktop renders at 120%. An explicit `FYNE_SCALE` launcher setting takes precedence; the picker explains when that override is active. Phones retain native sizing.

Close the dialog with **Close**, **Esc** or Android **Back**. Escape dismisses an open zoom list first. Closing Preferences restores the previous calculator focus.
