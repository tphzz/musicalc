# Tempo to Delay

Delay times and modulation rates for every note value at a tempo: what to set a delay, an LFO or a gate to so it follows the song.

## When to use it

- Setting a delay plug-in or hardware delay to quarter notes, dotted eighths or triplets.
- Syncing an LFO, tremolo, auto-pan or filter sweep to the tempo by hand.
- Setting pre-delay, gate or release times that breathe with the song.

## How it works

One beat (a quarter note) lasts **60000 ÷ BPM** milliseconds: 500 ms at 120 BPM. Every other note value is a multiple of that: a half note is twice as long, an eighth note half as long.

- **Dotted** notes are 1.5 times as long as straight ones.
- **Triplets** are two thirds as long: three of them fit in the time of two.
- The **rate in Hz** is how often that note value repeats per second: 1000 ÷ the time in ms. That's what LFOs usually ask for.

## How to use it

1. **Enter the tempo** in BPM in the field at the top.
2. Or **tap it in**: tap the pad under the field along with the music. On a computer you can also click the pad, or press **T** while the tempo field isn't focused.
3. **Read the cards**: one card per note value, from 1/1 (a whole note) to 1/64.

## Reading the results

- Each card shows the **Straight**, **Dotted** and **Triplet** time in milliseconds.
- The value on the right of each line is the matching **rate in Hz**.
- While you tap, each tap writes the measured tempo into the field, so all cards follow live. A pause of 2 seconds ends the measurement.

## In practice

- **Classic delays:** a dotted eighth (1/8 dotted) is the "galloping" guitar delay; a quarter note gives a clear, rhythmic echo; an eighth-note triplet swings.
- **Slap-back** delays are usually shorter than any note value (60–120 ms) and don't need to follow the tempo.
- **Modulation:** an LFO at the 1/2 or 1/1 rate gives slow, musical sweeps; 1/16 gives fast tremolo or auto-pan.
- If your plug-in has tempo sync, use that instead; these values are for gear and plug-ins that only take milliseconds or Hz.

## Examples

At 120 BPM:

| Note | Straight | Dotted | Triplet |
|---|---|---|---|
| 1/4 | 500 ms (2 Hz) | 750 ms | 333.33 ms |
| 1/8 | 250 ms (4 Hz) | 375 ms | 166.67 ms |

## Good to know

- The **Tap Tempo** calculator has a larger tap area and also shows how steady your taps were; its **Delay** button sends the tempo straight to this calculator.
