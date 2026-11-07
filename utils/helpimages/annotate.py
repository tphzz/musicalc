"""Crops screenshots of the calculators for the help pages and adds numbered callouts.

The screenshots are rendered by the capture program, 400 points wide at scale 2 (800 px), with the
examples of the help pages entered. Pass the folder it wrote them to:

    go run ./utils/helpimages/capture <capture folder>
    python utils/helpimages/annotate.py <capture folder>

Captures: cap-alignment.png (Alignment Delay, snare example), cap-sum.png (Two-Mic Sum, kick
4 m), cap-sum-aligned.png (the same with Mic A delayed by 11.65 ms, as Align sets it), cap-31.png
(3:1 Rule, Singer A 1 / 3 m, +6 dB, 10 dB target). Coordinates below are in the captures' pixels.
"""
import os
import sys

from PIL import Image, ImageDraw

from diagrams import AMBER, CARD, TEXT, font

OUT = os.path.join("internal", "ui", "help", "images")
WIDTH = 640  # 2x the shown width

# image: (capture, crop box, callouts [(x, y, label)])
IMAGES = {
    "alignment-cards.png": ("cap-alignment.png", (0, 260, 800, 1050), [
        (540, 306, "a delay to add, in ms and samples"),
        (770, 700, "a stereo pair with different delays"),
        (540, 848, "the mic at the reference distance"),
        (400, 980, "farther than the reference: move it earlier"),
    ]),
    "sum-graph.png": ("cap-sum.png", (0, 650, 800, 1135), [
        (299, 757, "+6 dB: both mics add up fully"),
        (163, 1023, "a dip: the mics cancel each other"),
        (281, 946, "shaded: the range of interest"),
        (231, 799, "the cursor: the level at 60 Hz"),
    ]),
    "sum-aligned.png": ("cap-sum-aligned.png", (0, 505, 800, 1212), [
        (457, 833, "after Align: +6 dB across the graph"),
        (380, 1045, "dashed: the response before"),
        (720, 679, "the 2.45 µs the 0.01 ms steps leave"),
    ]),
    "three-to-one-results.png": ("cap-31.png", (0, 378, 800, 1135), [
        (420, 417, "the distance check: geometry only"),
        (440, 588, "the spill against the target"),
        (440, 881, "what to change to reach it"),
    ]),
}


def marker(d, x, y, n, f):
    r = 17
    d.ellipse((x - r, y - r, x + r, y + r), fill=AMBER, outline=(20, 20, 20), width=3)
    w = d.textlength(str(n), font=f)
    d.text((x - w / 2, y - 15), str(n), font=f, fill=(20, 20, 20))


def annotate(folder, name, capture, box, callouts):
    img = Image.open(os.path.join(folder, capture)).convert("RGBA").crop(box)
    d = ImageDraw.Draw(img)
    fm = font(True, 22)
    for i, (x, y, _) in enumerate(callouts, 1):
        marker(d, x - box[0], y - box[1], i, fm)
    # Legend below the screenshot, on the card colour
    scale = WIDTH / img.width
    img = img.resize((WIDTH, round(img.height * scale)), Image.LANCZOS)
    fl = font(size=21)
    line = 34
    legend = Image.new("RGBA", (WIDTH, 16 + line * len(callouts)), CARD + (255,))
    ld = ImageDraw.Draw(legend)
    for i, (_, _, label) in enumerate(callouts, 1):
        y = 10 + line * (i - 1)
        marker(ld, 26, y + 15, i, font(True, 18))
        ld.text((52, y + 2), label, font=fl, fill=TEXT)
    out = Image.new("RGBA", (WIDTH, img.height + legend.height), (0, 0, 0, 0))
    out.paste(img, (0, 0))
    out.paste(legend, (0, img.height))
    # Rounded corners, like the app's cards
    mask = Image.new("L", out.size, 0)
    ImageDraw.Draw(mask).rounded_rectangle((0, 0, out.width - 1, out.height - 1), radius=20, fill=255)
    final = Image.new("RGBA", out.size, (0, 0, 0, 0))
    final.paste(out, (0, 0), mask)
    final.save(os.path.join(OUT, name), optimize=True)


if __name__ == "__main__":
    if len(sys.argv) != 2:
        raise SystemExit(__doc__)
    for name, (capture, box, callouts) in IMAGES.items():
        annotate(sys.argv[1], name, capture, box, callouts)
    print("screenshots written to", OUT)
