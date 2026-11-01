# Tempo zu Delay

Verzögerungszeiten und Modulationsfrequenzen für alle Notenwerte eines Tempos: passende Einstellungen für Delay, LFO oder Gate im Takt des Songs.

## Einsatz

- Viertel, punktierte Achtel oder Triolen auf einem Delay einstellen.
- LFO, Tremolo, Auto-Pan oder Filterbewegung von Hand mit dem Tempo synchronisieren.
- Reverb-Pre-Delay, Gate oder Release auf einen Notenwert abstimmen.

## Funktionsweise

Ein Schlag als Viertelnote dauert **60000 ÷ BPM** Millisekunden: 500 ms bei 120 BPM. Andere Notenwerte sind Vielfache davon: Eine halbe Note dauert doppelt, eine Achtel halb so lang.

- **Punktierte** Noten dauern das 1.5-Fache gerader Noten.
- **Triolen** dauern zwei Drittel: Drei passen in die Zeit von zwei geraden Noten.
- Die **Frequenz in Hz** gibt Wiederholungen pro Sekunde an: 1000 ÷ Dauer in ms. LFOs erwarten meist diesen Wert.

## Bedienung

1. Oben das **Tempo** in BPM eingeben.
2. Oder auf dem Feld darunter **im Takt tippen**. Am Computer funktioniert auch Klicken oder **T**, solange das Tempofeld keinen Fokus hat.
3. Die Karten für Notenwerte von 1/1 (ganze Note) bis 1/64 ablesen.

## Ergebnisse

- Jede Karte zeigt die Dauer **Gerade**, **Punktiert** und **Triole** in Millisekunden.
- Rechts steht jeweils die passende **Frequenz in Hz**.
- Jeder Tap schreibt das gemessene Tempo ins Eingabefeld; alle Karten folgen sofort. Eine Pause von 2 Sekunden beendet die Messung.

## Praxis

- **Klassische Delays:** Punktierte Achtel ergeben das galoppierende Gitarren-Delay, Viertel ein klares rhythmisches Echo und Achteltriolen einen triolischen Puls.
- **Slapback** liegt meist unter den Notenwerten, etwa bei 60–120 ms, und muss nicht temposynchron sein.
- **Modulation:** 1/2 und 1/1 ergeben langsame Bewegungen; 1/16 schnelles Tremolo oder Auto-Pan.
- Falls das Plug-in Temposynchronisation bietet, diese verwenden. Diese Werte sind für Geräte, die nur Millisekunden oder Hz annehmen.

## Beispiele

Bei 120 BPM:

| Note | Gerade | Punktiert | Triole |
|---|---|---|---|
| 1/4 | 500 ms (2 Hz) | 750 ms | 333.33 ms |
| 1/8 | 250 ms (4 Hz) | 375 ms | 166.67 ms |

## Hinweise

- Die Zeiten setzen ein gleichbleibendes Tempo voraus. Bei wechselndem Tempo passen sie nur zum eingegebenen Wert.
- **Tap-Tempo** hat ein größeres Tippfeld und zeigt die Gleichmäßigkeit der Taps. **Delay** überträgt das Tempo direkt in diesen Rechner.
