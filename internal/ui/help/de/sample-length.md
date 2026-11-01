# Samplelänge

Länge eines Loops in Samples und Millisekunden sowie das dazu passende Tempo. Tempo oder eine der beiden Längen eingeben; die anderen Werte folgen.

## Einsatz

- Einen Loop im Sampler oder Editor auf eine genaue Anzahl von Schlägen schneiden.
- Aus der Sampleanzahl eines Loops sein Tempo bestimmen.
- Die Sampleanzahl eines Taktes bei gegebenem Tempo und gegebener Samplerate prüfen.

## Funktionsweise

Ein Schlag dauert 60 ÷ BPM Sekunden. Ein Loop dauert daher Schläge × 60 ÷ BPM Sekunden. Multipliziert mit der Samplerate ergibt das seine Sampleanzahl. Die Rechnung funktioniert vom Tempo, von der Sampleanzahl oder von der Millisekundendauer aus.

## Bedienung

1. **Samplerate** wählen, von 8000 bis 192000 Hz, oder einen eigenen Wert eingeben.
2. **Schläge** einstellen: 4 entspricht einem 4/4-Takt, 16 vier solchen Takten.
3. **Tempo**, **Länge in Samples** oder **Länge in ms** eingeben. Die beiden anderen Werte folgen.

Ein Wechsel von Samplerate oder Schlaganzahl behält das Tempo und berechnet die Längen neu. **Reset** stellt 44100 Hz, 4 Schläge und 120 BPM wieder her.

## Ergebnisse

- **Tempo:** BPM-Wert, bei dem der Loop genau die eingegebene Anzahl Schläge lang ist.
- **Länge in Samples:** Sampleanzahl, etwa für Start- und Endpunkte eines Samplers.
- **Länge in ms:** Dauer in Millisekunden.

## Praxis

- **Loop-Tempo finden:** Zuerst Samplerate und gehörte Schlaganzahl einstellen. Dann die Sampleanzahl aus dem Editor übernehmen und das Tempo ablesen.
- **Loop schneiden:** Tempo und Schläge einstellen und bei der angezeigten Sampleanzahl schneiden.
- Die Schläge sorgfältig zählen. Ein halb oder doppelt so hohes Ergebnis weist oft auf eine um Faktor zwei falsche Schlaganzahl hin.

## Beispiele

- 88200 Samples bei 44.1 kHz dauern 2000 ms. Bei 4 Schlägen entspricht das **120 BPM**.
- Vier Takte mit 16 Schlägen bei 140 BPM ergeben **302400 Samples** beziehungsweise 6857.14 ms bei 44.1 kHz.

## Hinweise

- Ein Tempo ergibt selten eine ganzzahlige Sampleanzahl. Der Rechner zeigt ganze Samples und verwirft den Bruchteil.
