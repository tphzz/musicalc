# Tempoänderung

Wie eine Tempoänderung Tonhöhe und Länge von Audio beeinflusst und welches Tempo oder welche Zeitdehnung zu einer Transposition gehört.

## Einsatz

- Ein Sample oder einen Loop wie auf einem Sampler oder Plattenspieler beschleunigen: Tempo und Tonhöhe ändern sich gemeinsam.
- Die Transposition eines bei anderem Tempo abgespielten Samples ermitteln.
- Den Prozentwert für ein Time-Stretching-Werkzeug bestimmen.

## Funktionsweise

Wird Audio ohne Time-Stretching schneller abgespielt, steigt die Tonhöhe; bei langsamerer Wiedergabe sinkt sie. Doppeltes Tempo erhöht die Tonhöhe um eine Oktave, also 12 Halbtöne.

- **Transposition** in Halbtönen = 12 × log2(neues Tempo ÷ Ausgangstempo).
- **Zeitdehnung %** ist die neue **Länge** relativ zum Original: Ausgangstempo ÷ neues Tempo × 100. Über 100 % wird Audio länger und langsamer, unter 100 % kürzer und schneller.
- **Tempoänderung** zeigt die Änderung in Prozent.

## Bedienung

1. Das **Ausgangstempo** des Audiomaterials einstellen.
2. Einen der folgenden Werte ändern; die anderen folgen:
   - **Neues Tempo**: gewünschtes Tempo.
   - **Zeitdehnung %**: gewünschte Länge relativ zum Original.
   - **Transposition Halbtöne** und **Transposition Cent**: gewünschte Tonhöhenänderung.
3. **Tauschen** vertauscht Ausgangs- und neues Tempo. **Reset** stellt 120 → 100 BPM wieder her.

## Ergebnisse

- **Transposition Halbtöne / Cent:** Ganze Halbtöne plus 0 bis 99 Cent; 100 Cent sind ein Halbton, etwa +3 Halbtöne +70 Cent.
- **Halbtöne / Cent (50¢n):** Dieselbe Änderung relativ zum nächsten Halbton mit −50 bis +50 Cent Rest, etwa +4 Halbtöne −30 Cent.
- **Zeitdehnung %:** 100 % erhält die Länge; 200 % verdoppelt sie und halbiert das Tempo; 50 % halbiert sie und verdoppelt das Tempo.
- **Tempoänderung:** Wie viel schneller (+) oder langsamer (−) das neue Tempo ist.

## Praxis

- Einen **Loop auf Projekttempo bringen**: Loop- und Projekttempo eingeben und die gezeigte Transposition am Sampler einstellen.
- **Kleine Änderungen** um wenige BPM liegen unter einem Halbton: bei Drums oft unauffällig, bei tonalem Material hörbar.
- Soll die **Tonhöhe erhalten bleiben**, ein Time-Stretching-Werkzeug mit dem Wert **Zeitdehnung %** verwenden.

## Beispiele

- 140 BPM auf 170 BPM: **+3.36 Halbtöne** (+3 Halbtöne, +36 Cent), Zeitdehnung **82.35 %**.
- **7 Halbtöne** höher entspricht ungefähr dem 1.5-fachen Tempo: 120 BPM werden **179.8 BPM**, Zeitdehnung 66.74 %.

## Hinweise

- Beide Tempi müssen im Bereich 5–999,99 BPM liegen; außerhalb bleiben die Ergebnisse leer. Die Transposition ist so begrenzt, dass das neue Tempo in diesem Bereich bleibt.
- Der Rechner beschreibt die gekoppelte Änderung von Tempo und Tonhöhe. Modernes Time-Stretching erhält die Tonhöhe; dafür ist nur **Zeitdehnung %** relevant.
