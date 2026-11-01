# Tap-Tempo

Das Tempo eines Songs durch Mittippen bestimmen.

## Einsatz

- Das Tempo eines Referenztracks, einer Liveband oder eines Samples finden.
- Schnell ein Tempo für Delays bestimmen; siehe **Tempo zu Delay**.

## Funktionsweise

Jeder Tap markiert einen Schlag. Zwischen zwei Taps ergibt sich das Tempo als 60000 ÷ Abstand in Millisekunden. Gemittelt werden bis zu 8 Intervalle, also 9 Taps. Mehr Taps gleichen kleine Timingfehler aus; anschließend folgt der Mittelwert den jüngsten Taps. Taps mit weniger als 150 ms Abstand zählen als versehentlicher Doppeltap und werden zusammengefasst. Eine Pause von 2 Sekunden beendet die Messung.

## Bedienung

1. Auf die **große Fläche im Takt tippen**. Am Computer geht auch Klicken, **Leertaste** oder **T**.
2. Mindestens 4 bis 8 Schläge weitertippen. Ab dem zweiten Tap erscheint ein Tempo.
3. **2 Sekunden pausieren**, um die Messung zu beenden. Das Tempo bleibt sichtbar; der nächste Tap beginnt eine neue Messung.

## Ergebnisse

- Die große Zahl zeigt das **Tempo** in BPM.
- Darunter stehen die **Tap-Anzahl**, die **Schwankung** (± BPM) und die Dauer einer **Viertelnote**.
- **±** beschreibt die Streuung einzelner Intervalle um den Mittelwert. Das misst auch dein Timing: Ein kleiner Wert bedeutet gleichmäßige Taps, nicht automatisch den richtigen Grundschlag.

## Praxis

- Auf **Viertelnoten** tippen, also auf die gezählten Grundschläge, statt auf jede Achtel.
- **Runden** wechselt zwischen zwei Nachkommastellen und ganzen BPM. Produktionen mit Klick haben oft ein ganzzahliges Tempo; Liveaufnahmen häufig nicht.
- **Delay** überträgt das Tempo an **Tempo zu Delay** und öffnet diesen Rechner.
- **Reset** löscht die Messung; am Computer funktioniert auch **Esc**.

## Beispiele

- 500 ms zwischen Taps ergeben **120 BPM**, eine Viertelnote dauert 500 ms.
- 600 ms zwischen Taps ergeben **100 BPM**.

## Hinweise

- Ist das Ergebnis halb oder doppelt so groß wie erwartet, wurden nur jeder zweite Schlag oder Achtelnoten getippt. Den Wert entsprechend verdoppeln oder halbieren.
- Das Ergebnis ist der Mittelwert der letzten Taps und folgt Tempoänderungen im Song nur begrenzt. Bei 120 BPM über 8 Intervalle ergibt ein letzter Tap mit 10 ms Verspätung etwa 119.7 BPM.
