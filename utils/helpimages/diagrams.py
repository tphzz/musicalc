"""Draws the concept diagrams of the help pages into internal/ui/help/images.

Run from the repository root:  python utils/helpimages/diagrams.py

The diagrams are drawn at twice the size they are shown at (640 px wide, shown 320 px wide), on the
app's card colour, so they look the same in the light and the dark theme. Text uses Noto Sans, the
font Fyne ships, from the Go module cache.
"""
import cmath
import glob
import math
import os

from PIL import Image, ImageDraw, ImageFont

OUT = os.path.join("internal", "ui", "help", "images")

# Colours of the app's dark theme
CARD = (40, 41, 46)
TEXT = (232, 232, 235)
MUTED = (150, 152, 158)
GRID = (75, 77, 84)
BLUE = (36, 124, 255)
AMBER = (255, 152, 0)
GREEN = (76, 200, 90)

W = 640  # image width at 2x


def font(bold=False, size=22):
    """Noto Sans from the Fyne module in the Go module cache"""
    pattern = os.path.join(os.path.expanduser("~"), "go", "pkg", "mod", "fyne.io", "fyne", "v2@*", "theme", "font",
                           "NotoSans-Bold.ttf" if bold else "NotoSans-Regular.ttf")
    paths = sorted(glob.glob(pattern))
    if not paths:
        raise SystemExit("Noto Sans not found; run go mod download first")
    return ImageFont.truetype(paths[-1], size)


def card(height):
    """A transparent image with a rounded card, like the app's cards"""
    img = Image.new("RGBA", (W, height), (0, 0, 0, 0))
    d = ImageDraw.Draw(img)
    d.rounded_rectangle((0, 0, W - 1, height - 1), radius=20, fill=CARD)
    return img, d


def centered(d, xy, text, f, fill=TEXT):
    w = d.textlength(text, font=f)
    d.text((xy[0] - w / 2, xy[1]), text, font=f, fill=fill)


def arrow(d, a, b, fill, width=3, head=12):
    d.line((a, b), fill=fill, width=width)
    ang = math.atan2(b[1] - a[1], b[0] - a[0])
    for s in (-1, 1):
        p = (b[0] - head * math.cos(ang + s * 0.45), b[1] - head * math.sin(ang + s * 0.45))
        d.line((b, p), fill=fill, width=width)


def mic(d, x, y, colour):
    """A small microphone: head and stem"""
    d.ellipse((x - 11, y - 11, x + 11, y + 11), fill=colour)
    d.line((x, y + 11, x, y + 30), fill=colour, width=4)


def alignment_setup():
    """Source, a close mic, an overhead and the room mic; when each mic hears the hit, before and after delay"""
    img, d = card(540)
    f, fb, fs = font(size=22), font(True, 22), font(size=19)
    d.text((28, 20), "One source, measured to every mic", font=fb, fill=TEXT)
    d.text((W - 150, 24), "not to scale", font=fs, fill=MUTED)
    y0 = 120
    sx = 60
    d.ellipse((sx - 26, y0 - 26, sx + 26, y0 + 26), outline=TEXT, width=4)
    centered(d, (sx, y0 + 34), "Snare", fs, MUTED)
    mics = [("Snare Top", 0.1, BLUE, 175), ("Overhead", 1.2, BLUE, 345), ("Room", 3.0, AMBER, 545)]
    for name, dist, colour, x in mics:
        mic(d, x, y0 - 14, colour)
        centered(d, (x, y0 + 30), name, fs, TEXT)
        centered(d, (x, y0 + 56), f"{dist:g} m" + (" (reference)" if colour == AMBER else ""), fs, MUTED)
    # When each mic hears the hit
    c = 331.3 + 0.606 * 20
    t_ref = 3.0 / c * 1000
    tx0, tx1 = 40, W - 40
    tx = lambda ms: tx0 + (tx1 - tx0) * ms / 10
    rows = [(305, "Without delay: the hit arrives at different times"), (450, "With the delays added: all at once")]
    for row, (yy, title) in enumerate(rows):
        d.text((28, yy - 82), title, font=f, fill=TEXT)
        d.line((tx0, yy, tx1, yy), fill=GRID, width=2)
        for ms in (0, 5, 10):
            centered(d, (tx(ms), yy + 18), f"{ms} ms", fs, MUTED)
        if row == 0:
            for name, dist, colour, _ in mics:
                t = dist / c * 1000
                d.line((tx(t), yy - 18, tx(t), yy + 12), fill=colour, width=5)
                d.text((tx(t) + 8, yy - 44), f"{t:.2f}", font=fs, fill=colour)
        else:
            for i, (name, dist, colour, _) in enumerate(mics):
                d.line((tx(t_ref) - 6 + 6 * i, yy - 18, tx(t_ref) - 6 + 6 * i, yy + 12), fill=colour, width=4)
            d.text((tx(t_ref) - 210, yy - 46), f"all at {t_ref:.2f} ms", font=fs, fill=TEXT)
    img.save(os.path.join(OUT, "alignment-setup.png"))


def comb_filter():
    """A delayed copy (time) and the comb it makes (frequency)"""
    img, d = card(500 + 30)
    f, fb, fs = font(size=22), font(True, 22), font(size=19)
    d.text((28, 20), "A delayed copy …", font=fb, fill=TEXT)
    # Time: Mic A's click and Mic B's later click
    y = 120
    d.line((40, y, W - 40, y), fill=GRID, width=2)
    xa, xb = 120, 300
    d.line((xa, y, xa, y - 60), fill=BLUE, width=6)
    d.line((xb, y, xb, y - 60), fill=AMBER, width=6)
    centered(d, (xa, y + 8), "Mic A", fs, BLUE)
    centered(d, (xb, y + 8), "Mic B, later", fs, AMBER)
    arrow(d, (xa + 8, y - 30), (xb - 8, y - 30), MUTED, width=2)
    centered(d, ((xa + xb) / 2, y - 62), "delay Δt", fs, MUTED)
    # Frequency: |1 + e^(-jωΔt)| for Δt = 1 ms, 0–5 kHz
    d.text((28, 190), "… makes a comb: peaks and dips over frequency", font=fb, fill=TEXT)
    top, bottom, left, right = 280, 450, 60, W - 30
    top_db, bottom_db = 8, -24
    yy = lambda db: top + (top_db - max(bottom_db, min(top_db, db))) * (bottom - top) / (top_db - bottom_db)
    for db, lab in ((6, "+6"), (0, "0"), (-12, "-12")):
        d.line((left, yy(db), right, yy(db)), fill=GRID, width=1)
        d.text((left - 44, yy(db) - 13), lab, font=fs, fill=MUTED)
    pts = []
    for i in range(0, 801):
        fhz = 5000 * i / 800
        v = abs(1 + cmath.exp(-2j * math.pi * fhz * 0.001))
        db = 20 * math.log10(max(v, 1e-6))
        pts.append((left + (right - left) * i / 800, yy(db)))
    d.line(pts, fill=BLUE, width=4, joint="curve")
    for k, lab in ((0, "0"), (1, "1 kHz"), (2, "2"), (3, "3"), (4, "4"), (5, "5 kHz")):
        centered(d, (left + (right - left) * k / 5, bottom + 8), lab, fs, MUTED)
    d.text((left, top - 40), "peaks (+6 dB) every 1 / Δt, dips in between", font=fs, fill=MUTED)
    img.save(os.path.join(OUT, "comb-filter.png"))


def three_to_one():
    """Source A, its own mic at d and the neighbouring mic at 3 × d"""
    img, d = card(330)
    f, fb, fs = font(size=22), font(True, 22), font(size=19)
    d.text((28, 20), "One source, two mics", font=fb, fill=TEXT)
    y = 150
    sx, own, neigh = 70, 70 + 150, 70 + 450
    d.ellipse((sx - 24, y - 24, sx + 24, y + 24), outline=TEXT, width=4)
    centered(d, (sx, y + 32), "Singer A", fs, MUTED)
    mic(d, own, y - 12, BLUE)
    centered(d, (own, y + 32), "its mic", fs, BLUE)
    mic(d, neigh, y - 12, AMBER)
    centered(d, (neigh, y + 32), "neighbour mic", fs, AMBER)
    arrow(d, (sx + 30, y - 52), (own - 16, y - 52), BLUE)
    centered(d, ((sx + own) / 2, y - 88), "d", fb, BLUE)
    d.line((own + 18, y - 52, neigh - 16, y - 52), fill=AMBER, width=3)
    arrow(d, (neigh - 60, y - 52), (neigh - 16, y - 52), AMBER)
    centered(d, ((own + neigh) / 2, y - 88), "at least 3 × d", fb, AMBER)
    d.text((28, 236), "Three times the distance: Singer A's spill in", font=f, fill=TEXT)
    d.text((28, 268), "the neighbour mic is about 9.5 dB lower.", font=f, fill=TEXT)
    img.save(os.path.join(OUT, "three-to-one.png"))


if __name__ == "__main__":
    os.makedirs(OUT, exist_ok=True)
    alignment_setup()
    comb_filter()
    three_to_one()
    print("diagrams written to", OUT)
