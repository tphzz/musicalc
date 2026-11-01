# Timecode

Timecodes addieren und subtrahieren und ihre Frameanzahl bei verschiedenen Bildraten bestimmen.

## Einsatz

- Längen von Clips, Cues oder Filmrollen addieren.
- Abstand zwischen zwei Timecodes bestimmen, etwa Start und Ende eines Cues.
- Timecode für framebasierte Werkzeuge in eine Frameanzahl umrechnen.
- Frameanzahl eines Timecodes bei einer anderen Bildrate prüfen.

## Funktionsweise

`HH:MM:SS:FF` zählt Stunden, Minuten, Sekunden und Frames. Die Frameanzahl hängt von der Bildrate ab: Eine Sekunde umfasst bei 25 fps 25 Frames, bei 30 fps 30. Der Rechner wandelt beide Eingaben in Frameanzahlen um, addiert oder subtrahiert sie und stellt das Ergebnis wieder als Timecode dar.

**Drop-Frame** (29.97 df) hält NTSC-Timecode mit der Uhrzeit synchron. Das Video läuft mit 29.97 statt 30 Bildern pro Sekunde. Deshalb werden am Anfang jeder Minute außer jeder zehnten die Framenummern 00 und 01 ausgelassen. Der Rechner berücksichtigt diese ausgelassenen Nummern.

## Bedienung

1. **Bildrate wählen:** 23.976, 24, 25, 29.94, 29.97, 29.97 Drop-Frame (df), 30, 50, 59.94 oder 60 fps.
2. **Ersten Timecode eingeben.** Ziffern werden von rechts aufgefüllt: erst Frames, dann Sekunden, Minuten und Stunden.
3. Den **zweiten Timecode** ebenso eingeben.
4. Mit **+** addieren oder mit **-** den zweiten vom ersten Timecode abziehen. Das Ergebnis steht anschließend im ersten Feld für die nächste Rechnung.

| Taste | Wirkung |
|---|---|
| Ziffern | Von rechts auffüllen: Frames, Sekunden, Minuten, Stunden |
| **.** (Punkt) | Fügt `00` ein; `1.` bedeutet etwa 1 Sekunde |
| Rücktaste | Entfernt die letzte Ziffer |

Ein leeres Feld zählt als 0.

## Ergebnisse

- Neben jedem Timecode steht seine **Frameanzahl** bei der aktuellen Bildrate, etwa `(62254f @ 30)`.
- Der **Verlauf** zeigt Rechnungen und Bildratenwechsel mit Frameanzahlen, neueste zuerst. Er hält die letzten 100 Einträge. Am Computer lassen sie sich markieren und kopieren.
- Ein **Bildratenwechsel** behält den geschriebenen Timecode bei, zählt dessen Frames neu und protokolliert die Umrechnung.

## Praxis

- **Mehrere Clips addieren:** ersten und zweiten Clip eingeben, **+** drücken, dritten eingeben, erneut **+** usw. Die Summe bleibt im ersten Feld.
- **Cue-Länge:** Ende zuerst und Start danach eingeben, dann **-** drücken.
- Für andere Werkzeuge zuerst die **Bildrate des Projekts** einstellen und dann die Frameanzahl ablesen.
- **Reset** leert beide Eingaben und den Verlauf. **Leeren** leert nur den Verlauf.

## Beispiele

- Bei 30 fps: `00:22:22:22 + 00:12:12:12 = 00:34:35:04`, also 62254 Frames.
- Derselbe Timecode `00:34:35:04` bei 25 fps: 51879 Frames.
- Bei 29.97 df entspricht `00:10:00:00` 17982 Frames: In zehn Minuten wurden 18 Nummern ausgelassen.

## Hinweise

- Timecode nummeriert ganze Frames. 23.976, 29.97 und 59.94 fps zählen daher wie 24, 30 und 60 fps. Nur **29.97 df** lässt Nummern aus. **29.94** zählt wie 29.97.
- Der Rechner arbeitet mit Timecode und Frameanzahlen, nicht mit Uhrzeit. Er zeigt nicht die tatsächliche Sekundendauer bei NTSC-Bildraten.
- Ein Bildratenwechsel zählt nur die Frames des geschriebenen Timecodes neu. Er überträgt weder eine Zeitposition noch eine Dauer auf eine andere Bildrate.
