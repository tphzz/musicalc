# Zwei-Mikrofon-Summe

Was beim Mischen zweier Mikrofone derselben Quelle in unterschiedlichen Abständen passiert und wie Verzögerung, Pegel oder Polarität das Ergebnis ändern.

## Einsatz

- Bassdrum, Snare oder Verstärker mit Nah- und Raummikrofon: Wird der Bass beim Mischen dünner?
- Zwischen Laufzeitausgleich und Polaritätsumkehr entscheiden.
- Hohlen oder phasigen Klang und die betroffenen Frequenzen verstehen.

## Funktionsweise

Das weiter entfernte Mikrofon hört denselben Schall später. Die Mischung fügt eine **verzögerte Kopie** hinzu. Bei manchen Frequenzen addieren sich die Anteile: **Verstärkung**, bei gleichem Pegel bis +6 dB. Bei anderen liegen sie gegensinnig und erzeugen **Auslöschung**. Dieses regelmäßig wiederkehrende Muster heißt **Kammfilterung**.

![Verzögerte Kopie und entstehender Kammfilter](../images/comb-filter.png)

*1 ms Verzögerung: Spitzen alle 1000 Hz, Senken jeweils dazwischen. Die Abbildung verwendet englische Beschriftungen.*

Das Muster folgt dem Zeitunterschied Δt:

- **Senken** bei 1 ÷ (2 × Δt), danach alle 1 ÷ Δt. Bei 11.65 ms etwa 42.9 Hz, 128.8 Hz, 214.6 Hz …
- **Spitzen** bei Vielfachen von 1 ÷ Δt: 85.9 Hz, 171.7 Hz …
- **Pegel und Polarität** ändern Tiefe und Position. Ein leiseres B macht Senken und Spitzen flacher. Umgekehrte Polarität vertauscht ihre Positionen.

Das Diagramm zeigt den **Summenpegel relativ zu Mikrofon A allein**. 0 dB entspricht A allein; +6 dB einer vollständigen Addition beider gleich lauten Anteile.

## Bedienung

1. Die Karte **Aufbau** prüfen:
   - **Schallquelle:** typischer Frequenzbereich mit Markierung, etwa Bassdrum 50–100 Hz bei 60 Hz oder Becken 3–12 kHz. **Bereich** und **Frequenz** sind frei änderbar; die Auswahl zeigt danach **Benutzerdefiniert**.
   - **Temperatur**, **Einheiten** und **Samplerate** einstellen.
2. **Mikrofon B gegenüber Mikrofon A** beschreiben:
   - **Weiter entfernt um:** zusätzlicher Abstand von B als Entfernung oder gemessene Zeit in **ms**. Ein Einheitenwechsel rechnet um. Negativ bedeutet, dass B näher steht.
   - **Pegel:** B relativ zu A in dB; 0 bedeutet gleich laut.
   - **Polarität:** Normal oder Invertiert wie am Kanal.
   - **Verzögerung auf A:** auf A angewendete Verzögerung in ms.
3. Änderungen ausprobieren. **Ausrichten** setzt den Laufzeitausgleich für den Direktschall. Pegel und Polarität ändern und das Diagramm beobachten.

## Ergebnisse

- **Ankunft:** etwa „Mikrofon B kommt 11.65 ms später an“, zusätzlich als Sampleanzahl. Nach dem Ausrichten steht **Ausgerichtet**. **Reset** entfernt die Verzögerung.

![Frequenzgang mit nummerierten Elementen](../images/sum-graph.png)

*Die Abbildung verwendet englische Beschriftungen.*

- Das **Diagramm** zeigt den Summenpegel über der Frequenz. Der **schattierte Bereich** ist der gewählte Frequenzbereich. Die **Markierung** zeigt die eingestellte Frequenz. Darunter stehen deren Pegel sowie Minimum und Maximum im Bereich.
- Wird der Kammfilter zu fein für eine einzelne Linie, zeigt das Diagramm das **Pegelband**.
- **Erwartete Auslöschung und Verstärkung** listet die ersten drei Senken und Spitzen ab dem Diagrammbereich samt Pegel. Bei dichtem Kammfilter erscheint auch der Abstand.
- **Direktschall ausrichten** nennt den nötigen Ausgleich und das zu verzögernde Mikrofon. Ob es besser klingt, entscheidet das Gehör.

![Nach dem Ausrichten: flache +6 dB und ursprünglicher Verlauf als gestrichelte Linie](../images/sum-aligned.png)

- Bei Verzögerung auf A, anderem Pegel als 0 dB oder invertierter Polarität zeigt eine **gestrichelte Linie** den Ausgangszustand zum Vergleich: beide gleich laut, normale Polarität, keine zusätzliche Verzögerung.

## Praxis

1. Den zusätzlichen Abstand des zweiten Mikrofons messen oder schätzen. Wenn möglich stattdessen den **Zeitversatz** zwischen zwei aufgenommenen Schlägen in der DAW messen und in ms eingeben.
2. Den relevanten **Frequenzbereich** betrachten. Eine Senke etwa bei 60 Hz kann eine Bassdrum dünn klingen lassen.
3. Korrekturen in dieser Reihenfolge prüfen:
   - **Ausrichten:** Näheres Mikrofon verzögern. Im Modell verschwindet der Kammfilter, reale Raumreflexionen haben aber andere Laufzeiten.
   - **Polarität:** Umkehren von B vertauscht Senken und Spitzen. Eine Senke kann dadurch verschwinden und eine andere in den wichtigen Bereich rücken.
   - **Pegel:** Leiseres B macht alle Senken flacher, verringert aber seinen Beitrag.
4. **In Mono abhören**, Änderung ein- und ausschalten und dem Gehör mehr vertrauen als dem Diagramm.

## Beispiele

B 4 m weiter entfernt, 20 °C, 48 kHz, Bassdrum 50–100 Hz:

| Einstellung | Ergebnis |
|---|---|
| Gleicher Pegel, normale Polarität | B kommt 11.65 ms beziehungsweise 559 Samples später an. Bei 60 Hz entstehen +1.4 dB, im Bereich 50–100 Hz −5.8 bis +6.0 dB. |
| Ausrichten: A um 11.65 ms verzögern | Im Modell +6.0 dB bei jeder Frequenz. |
| B bei −6 dB, invertierte Polarität | Senken nur bis −6.0 dB, Spitzen +3.5 dB. Senken verschieben sich zu 85.9 · 171.7 Hz. |

## Hinweise

- Das einfache Modell verwendet dasselbe Signal mit Verzögerung, Pegel und Polarität. Raumreflexionen, Mikrofonfrequenzgänge und Richtwirkung fehlen; reale Verläufe sind unregelmäßiger.
- Die vorgeschlagene Ausrichtung ist eine **Hörentscheidung**. Der zeitliche Abstand eines Raummikrofons kann zum gewünschten Klang gehören.
- **Laufzeitausgleich** berechnet Verzögerungen für mehrere Mikrofone. Die **3:1-Regel** prüft das Übersprechen zu einem Nachbarmikrofon.
