# Note zu Frequenz

Die Frequenz jeder Note in der Standardstimmung oder einer von 45 weiteren Stimmungen.

## Einsatz

- Synthesizer, Trommel oder 808-Bassdrum auf die Tonart eines Songs stimmen.
- Die Frequenz einer Note für EQ, Filter oder Resonanz finden.
- Historische Temperaturen, reine Stimmung und andere Tonsysteme erkunden.

## Funktionsweise

In der Standardstimmung **12-TET** (gleichstufige Zwölftonteilung) ist jeder Halbton gleich groß: Die Frequenz steigt um den Faktor 2^(1/12), etwa 1.0595, und verdoppelt sich pro Oktave. Eine **Referenznote**, üblicherweise A = 440 Hz, legt alle übrigen Frequenzen fest.

Andere Stimmungen teilen die Oktave anders. Die App enthält sie als Scala-Skalen; eine Skalendatei nennt nur ihre Stufen über einer unbenannten ersten Stufe. Der **Grundton** legt fest, auf welcher Note diese erste Stufe liegt. Jede Stimmung beginnt auf ihrem eigenen Grundton, meist C (D bei Bayati, E bei Huzam); ein anderer Grundton kann gewählt werden. Die Referenznote behält dabei ihre Frequenz. Bei Kirnberger III auf C ist zum Beispiel C–E eine reine Terz, auf A dagegen A–C#. Bei Stimmungen mit 12 Stufen zeigt die Tabelle die Abweichung jeder Note von der gleichstufigen Stimmung in **Cent**; 100 Cent sind ein Halbton.

Stimmungen mit einer anderen Stufenzahl, etwa 5, 7 oder 13, passen nicht zu den zwölf Notennamen. Sie werden wie in Scala Taste für Taste verteilt: Stufe 0 liegt auf der Grundton-Taste auf oder unter der Referenz, jede folgende MIDI-Taste spielt die nächste Stufe. In einer 7-stufigen Skala ist die Taste „A3“ daher nur eine Taste, nicht unbedingt das musikalische A.

## Bedienung

1. **Referenz** mit Note und Frequenz einstellen. Standard ist A3 = 440 Hz in der C3-Konvention, bei der das mittlere C als C3 bezeichnet wird. Die Note wird bei Bedarf mit Kreuz und mit Oktave geschrieben, z. B. A3, C#4 oder C-2; die Frequenz kann 1 Hz bis 20 kHz betragen.
2. **Mittleres C** auf C3 oder C4 stellen. Das ändert nur die Oktavnummern.
3. **Stimmung wählen**, standardmäßig Gleichstufig · 12-TET. Die 46 Stimmungen sind nach gleichstufigen, reinen, historischen und weiteren weltweiten Tonsystemen gruppiert. Die Auswahl ist markiert.
4. **Grundton** von C bis B wählen. Der eigene Grundton der Stimmung ist mit „Standard“ markiert; bei Wahl einer anderen Stimmung springt der Grundton auf deren Grundton zurück.
5. Die Tabelle ablesen.

## Ergebnisse

- 132 Noten von C-2 bis B8, bei der C4-Konvention C-1 bis B9, mit ihrer **Frequenz** in Hz.
- **Cent** zeigt die Abweichung zur gleichstufigen Stimmung; bei 12-TET ist sie 0.
- Bei Stimmungen ohne 12 Stufen heißt die erste Spalte **Stufe**: die Stufe und in Klammern die Periode (meist die Oktave), gezählt ab der Grundton-Taste auf oder unter der Referenz, z. B. „2 (1)“. Die Spalte Cent zeigt dann das Intervall der Stufe über dem Grundton ihrer Periode.
- **MIDI** zeigt zwei Nummerierungen: mittleres C = 60 sowie die von manchen Geräten verwendete Variante mit mittlerem C = 72. MIDI-Notennummern enden bei 127; höhere Noten zeigen „-“.
- Die Tabelle öffnet an der hervorgehobenen Referenznote. Die Grundtöne (Stufe 0) sind in der ersten Spalte fett; mit den Standardwerten ist das jedes C.
- Solange Referenznote oder Frequenz nicht verwendbar sind (z. B. „A“ ohne Oktave oder 0 Hz), zeigen die Spalten Frequenz und Cent „-“, bis die Eingabe korrigiert ist; ebenso die Spalte Stufe, die von der Referenztaste abhängt.

## Praxis

- **Bassdrum oder 808 stimmen:** Grundton des Songs und eine passende Oktave suchen, häufig 40–60 Hz. Die Bassdrum auf diese Frequenz stimmen.
- Für einen anderen **Kammerton** als 440 Hz, etwa 442 oder 432 Hz, die Referenzfrequenz ändern. Die gesamte Tabelle folgt.
- **Reset** stellt die Standardwerte wieder her.

## Beispiele

- Bei Standardstimmung ist A3 in der C3-Konvention 440 Hz; C3, das mittlere C, ist 261.63 Hz.
- Eine Oktave tiefer ist A2 220 Hz.
- Historisch · Kirnberger III auf C, A3 = 440 Hz: C3 ist 263.18 Hz (+10.26 Cent), E3 ist 328.98 Hz, genau 5/4 über C3.
- Rein · Ptolemäus/Zarlino (7 Stufen) auf C, A3 = 440 Hz: Die Taste A3 ist Stufe 2 (1) mit +386.31 Cent, die Taste C3 (MIDI 60) ist der Grundton, Stufe 0 (0), mit 176.00 Hz.

## Hinweise

- C3 und C4 benennen dieselben Töne mit anderen Oktavnummern: A3 der einen Konvention ist A4 der anderen. Hersteller verwenden unterschiedliche Konventionen; die zum eigenen Gerät passende wählen.
- Die internationalen Notennamen bleiben erhalten: B wird auch in der deutschen Oberfläche als B bezeichnet.
- **Frequenz zu Note** führt die umgekehrte Rechnung aus.
