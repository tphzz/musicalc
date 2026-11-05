"""Crops screenshots of the calculators for the help pages and adds numbered callouts.

The screenshots are taken from the phone-width app at twice the size (FYNE_SCALE=2, 400 x 680
window), with the examples of the help pages entered. Pass the folder that holds them:

    python utils/helpimages/annotate.py <capture folder>

Expected captures: cap-alignment.png (Alignment Delay, snare example), cap-sum.png (Two-Mic Sum,
kick 4 m), cap-sum-aligned.png (the same after Align), cap-31.png (3:1 Rule, Singer A 1 / 3 m,
+6 dB, 10 dB target). Coordinates below are in the captures' pixels (800 px wide).
"""
import os
import sys

from PIL import Image, ImageDraw

from diagrams import AMBER, CARD, TEXT, font

OUT = os.path.join("internal", "ui", "help", "images")
WIDTH = 640  # 2x the shown width

# image: (capture, crop box, callouts [(x, y, label)])
IMAGES = {
    "alignment-cards.png": ("cap-alignment.png", (0, 333, 800, 1122), [
        (420, 388, "a delay to add, in ms and samples"),
        (770, 745, "a stereo pair with different delays"),
        (420, 902, "the mic at the reference distance"),
        (420, 1035, "farther than the reference: move it earlier"),
    ]),
    "sum-graph.png": ("cap-sum.png", (0, 733, 800, 1211), [
        (333, 836, "+6 dB: both mics add up fully"),
        (182, 1100, "a dip: the mics cancel each other"),
        (290, 1050, "shaded: the range of interest"),
        (231, 882, "the cursor: the level at 60 Hz"),
    ]),
    "sum-aligned.png": ("cap-sum-aligned.png", (0, 589, 800, 1244), [
        (520, 878, "after Align: +6 dB at every frequency"),
        (444, 1105, "dashed: the response before"),
    ]),
    "three-to-one-results.png": ("cap-31.png", (0, 455, 800, 1233), [
        (560, 496, "the distance check: geometry only"),
        (440, 667, "the spill against the target"),
        (440, 940, "what to change to reach it"),
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
