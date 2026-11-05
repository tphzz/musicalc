# Orientierung

MusiCalc gruppiert seine Rechner in vier Kategorien. Jeder beantwortet eine alltägliche Frage aus Musikproduktion und Aufnahme; Ergebnisse aktualisieren sich während der Eingabe.

## Einen Rechner finden

- Die **Menütaste** (☰) oben links wechselt zwischen **Zeit & Tempo**, **Frequenz & Tonhöhe**, **Analyse** und **Mehrmikrofon**.
- Die **Tableiste** unten wechselt zwischen den Rechnern der aktuellen Kategorie. Auf Android und iPhone zeigt sie nur Symbole.
- Die Kopfzeile zeigt die aktuelle Position, etwa `Zeit & Tempo / Tap-Tempo`.

| Kategorie | Rechner |
|---|---|
| Zeit & Tempo | Timecode, Tempo zu Delay, Tempoänderung, Tap-Tempo |
| Frequenz & Tonhöhe | Note zu Frequenz, Frequenz zu Note |
| Analyse | Samplelänge |
| Mehrmikrofon | Laufzeitausgleich, Zwei-Mikrofon-Summe, 3:1-Regel |

## Hilfe und Aktionen

- **⋮** oben rechts öffnet das Rechnermenü. Einige Rechner bieten eigene Aktionen, etwa Presets beim Laufzeitausgleich. **Hilfe** ist immer der letzte Eintrag.
- Die Hilfe gehört zum gerade verwendeten Rechner. **✕** oben rechts schließt sie.
- Am Computer öffnet **F1** die Hilfe; **Esc** schließt sie. Auf Android schließt die Zurück-Taste sie.

## App-Version

**Über MusiCalc** steht im Menü oben links. Auf macOS befindet sich der Eintrag stattdessen unter **MusiCalc > Über MusiCalc** in der Systemmenüleiste. Der Dialog zeigt die Version; paketierte Mobilversionen zeigen zusätzlich die Buildnummer in Klammern. **Schließen**, **Esc** oder Android Zurück schließen ihn, ohne Rechnerwerte zu verändern.

## Gemeinsames Verhalten

- **Werte bleiben erhalten.** Beim Wechsel zwischen Rechnern und beim Öffnen der Hilfe bleiben Eingaben erhalten. Nur die Mehrmikrofon-Rechner speichern sie auch über das Schließen der App hinweg.
- **Aufbau-Karten.** Die Mehrmikrofon-Rechner haben oben eine Karte **Aufbau**. Ihre Kopfzeile klappt sie zu einer weiterhin aussagekräftigen Zusammenfassung zusammen und wieder auf. Der Aufbau bleibt beim Schließen gespeichert. Zwei-Mikrofon-Summe und 3:1-Regel merken sich auch den geöffneten Zustand; beim Laufzeitausgleich ist die Karte zunächst nur ohne Mikrofone geöffnet.
- **Einheiten werden umgerechnet.** Abstände können in Metern oder Fuß, Temperaturen in °C oder °F eingegeben werden. Ein Einheitenwechsel rechnet Werte um: 4 m werden 13.12 ft und 22 °C werden 71.6 °F. Die physikalischen Ergebnisse bleiben gleich.
- **Auf Android und iPhone** verschwindet die Tableiste während der Bildschirmtastatur und kehrt anschließend zurück.
- **Notennamen und Zahlenformate** behalten ihre technische Bedeutung. Die internationale Notenbezeichnung B bleibt auch auf Deutsch B; Dezimalpunkt und Dezimalkomma werden wie bisher als Eingabe akzeptiert.

## Skalierung am Computer

MusiCalc verwendet die Skalierung des Betriebssystems für seine Darstellungsgröße. Unter Windows steht sie unter **Einstellungen > System > Anzeige > Skalierung** für den jeweiligen Monitor. Eine höhere oder breitere Auflösung allein wählt keine größere Skalierung.

## Einstellungen

Im Menü oben links **Einstellungen…** öffnen. Eine Trennlinie gruppiert Einstellungen und Über MusiCalc unterhalb der Rechnerkategorien.

**Sprache:** **System**, **English** oder **Deutsch** wählen. System ist voreingestellt: Bei einem deutschen Betriebssystem-Gebietsschema verwendet MusiCalc Deutsch, sonst oder bei fehlender Erkennung Englisch. Eine ausdrücklich gewählte Sprache hat Vorrang und bleibt gespeichert. Änderungen gelten sofort für Oberfläche und Hilfe; Eingaben, Ergebnisse, Verlauf und der ausgewählte Rechner bleiben erhalten. Native Betriebssystemelemente behalten die Betriebssystemsprache.

**Skalierung (Desktop):** **80%, 100% (Standard), 115%, 130% oder 200%** wählen. Der Wert passt MusiCalc relativ zur Desktop-Skalierung des Betriebssystems an und wird sofort gespeichert. Die gesamte Oberfläche ändert ihre Größe ohne Neustart. Fyne rundet die resultierende Skalierung auf 10%-Schritte: 115% auf einem Desktop mit 100% werden beispielsweise als 120% dargestellt. Eine ausdrücklich gesetzte Startvariable `FYNE_SCALE` hat Vorrang; in diesem Fall erklärt die deaktivierte Auswahl den Grund. Auf Smartphones bleibt die native Größe erhalten.

Den Dialog mit **Schließen**, **Esc** oder Android **Zurück** schließen. Escape schließt zunächst eine geöffnete Skalierungsliste. Danach erhält das vorherige Rechnerfeld seinen Fokus zurück.
