"""Removes the semi-transparent "SAMPLE" overlay from yugioh_cardgame_official_jpn Instagram art.

Usage: set INPUT and OUTPUT below, then run
    python remove_yugiohjpn_watermark.py

INPUT and OUTPUT are both image files or both directories.

The overlay is a fixed layer:
    shown = a * original + b
where a and b are per-pixel maps stored in watermark_model.npz
(fitted by fit_yugiohjpn_watermark.py on 1080 pixel wide posts).
Restoring is original = (shown - b) / a.

The overlay is not always at the same place, so each image is searched for it first:
the model's letter pattern is slid over the image at the scale image_width / 1080
(and slightly around it), and the best match gives the position and scale.
The model is then resized and moved onto the image, so the art itself is never resampled.
The thin outline left on letter edges is then blended into the art beside it, where the art is smooth.
The copyright footer under the letters is left as it is.
Images where the pattern is not found (sketches, photos, banners) are copied unchanged.
"""
import os
import shutil
from concurrent.futures import ProcessPoolExecutor

import cv2
import numpy as np


HERE = os.path.dirname(os.path.abspath(__file__))

# Both are image files, or both are directories (OUTPUT is created if missing).
INPUT = r"D:\tmp_process_MD_file\yugiohjpn_arts_renamed"
OUTPUT = r"D:\tmp_process_MD_file\yugiohjpn_arts_clean"

MODEL_PATH = os.path.join(HERE, "watermark_model.npz")
# Width of the images the model was fitted on.
MODEL_WIDTH = 1080
# Minimum kept opacity, avoids amplifying noise where the model is unsure.
MIN_A = 0.2
# How far the best match stands out from the rest of the match map, in standard deviations.
# On 308 posts, images with the overlay scored 19 or more and images without it 13 or less.
MIN_MATCH_SCORE = 16
# Scales searched, relative to image_width / MODEL_WIDTH:
# a coarse pass over the whole image, then a fine pass around the best match.
COARSE_SCALE_STEPS = np.arange(-0.03, 0.0301, 0.005)
# Only if the coarse pass finds nothing: some posts use a smaller overlay
# (1024x1024 posts with black side bars have it at about 0.64 of image_width / MODEL_WIDTH).
# A match fades within about 2% of the right scale, so the steps are 1.5% apart.
WIDE_SCALES = 0.5 * 1.015 ** np.arange(47)
FINE_SCALE_STEPS = np.arange(-0.004, 0.00401, 0.001)
# Model rows holding the "SAMPLE" letters, the top of the copyright footer starts below them.
PATTERN_ROWS = 290
# Model row where the letters' shadow has faded out, the copyright footer layer starts a few rows below.
LETTER_ROWS_END = 304
# Sigma of the blur removed before matching, keeps letter edges and drops smooth shading.
HIGHPASS_SIGMA = 4
# Model opacity change per pixel that marks a letter edge (the soft shadow changes about 0.02).
EDGE_GRADIENT = 0.06
# Sigma of the blur that fills a letter edge with the colors just beside it.
EDGE_FILL_SIGMA = 2
# Art detail beside a letter edge (mean absolute difference from a blur, 0 to 255):
# fully smoothed below SMOOTH_ART, untouched above BUSY_ART, where the outline is hidden anyway.
SMOOTH_ART, BUSY_ART = 3, 8
# Difference between an edge pixel and its fill (0 to 255): fully corrected below OUTLINE_MAX,
# untouched above ART_LINE_MIN, which is more likely an art line crossing the letter edge than the faint outline.
OUTLINE_MAX, ART_LINE_MIN = 12, 24


def load_model(path=MODEL_PATH):
    z = np.load(path)
    a, b = z["a"].astype(np.float32), z["b"].astype(np.float32)
    # What the overlay does to mid-gray art, the pattern searched for in each image.
    # Only the letters: the copyright footer below them is also on posts without the overlay.
    pattern = ((a - 1) * 128 + b).mean(2)[:PATTERN_ROWS]
    # The model's bottom rows are the copyright footer layer (a gray band and the top of the text),
    # fitted against clean art that has no footer, so restoring them smears the footer text.
    # The footer is not the watermark: fade the model to "no change" below the letters.
    fade = np.clip((LETTER_ROWS_END - np.arange(a.shape[0])) / 6, 0, 1)[:, None, None].astype(np.float32)
    return a * fade + (1 - fade), b * fade, pattern


def highpass(gray):
    return gray - cv2.GaussianBlur(gray, (0, 0), HIGHPASS_SIGMA)


def placement(scale, x, y):
    """Affine map from model pixels to image pixels, model top-left corner at (x, y)."""
    shift = 0.5 * scale - 0.5  # keeps pixel centers aligned when scaling
    return np.float32([[scale, 0, x + shift], [0, scale, y + shift]])


def scaled_pattern(pattern, scale):
    height, width = pattern.shape
    size = (int(np.ceil(width * scale)), int(np.ceil(height * scale)))
    return highpass(cv2.warpAffine(pattern, placement(scale, 0, 0), size, flags=cv2.INTER_LINEAR))


def subpixel(scores, x, y):
    """Refines an integer peak of the match map with a parabola fit along each axis."""
    def offset(left, mid, right):
        curve = left - 2 * mid + right
        return 0.0 if curve >= 0 else 0.5 * (left - right) / curve
    height, width = scores.shape
    dx = offset(scores[y, x - 1], scores[y, x], scores[y, x + 1]) if 0 < x < width - 1 else 0.0
    dy = offset(scores[y - 1, x], scores[y, x], scores[y + 1, x]) if 0 < y < height - 1 else 0.0
    return x + dx, y + dy


def best_match(gray, pattern, scales, height):
    """Returns (peak, match score, scale, x, y) of the best scale, or None if no scale fits."""
    best = None
    for scale in scales:
        template = scaled_pattern(pattern, scale)
        if template.shape[0] >= height or template.shape[1] > gray.shape[1]:
            continue
        scores = cv2.matchTemplate(gray, template, cv2.TM_CCOEFF_NORMED)
        _, peak, _, (x, y) = cv2.minMaxLoc(scores)
        if best is None or peak > best[0]:
            match_score = (peak - scores.mean()) / (scores.std() + 1e-6)
            best = (peak, match_score, scale, x, y)
    return best


def locate(img, pattern):
    """Returns (match score, scale, x, y) of the overlay's top-left corner.

    The overlay was found if match score >= MIN_MATCH_SCORE.
    """
    height, width = img.shape[:2]
    pad = int(0.03 * width)  # lets the overlay stick out a little on either side
    gray = highpass(cv2.cvtColor(img, cv2.COLOR_BGR2GRAY).astype(np.float32))
    gray = cv2.copyMakeBorder(gray, 0, 0, pad, pad, cv2.BORDER_CONSTANT, value=0)
    base = width / MODEL_WIDTH
    best = best_match(gray, pattern, base * (1 + COARSE_SCALE_STEPS), height)
    if best is None or best[1] < MIN_MATCH_SCORE:
        wide = best_match(gray, pattern, base * WIDE_SCALES, height)
        if wide is not None and (best is None or wide[1] > best[1]):
            best = wide
    if best is None:
        return 0.0, base, 0.0, 0.0
    _, match_score, coarse_scale, coarse_x, coarse_y = best
    if match_score < MIN_MATCH_SCORE:
        return match_score, coarse_scale, coarse_x - pad, coarse_y

    # Fine pass: a few pixels around the coarse match, at finer scale steps.
    margin = 8
    top, left = max(coarse_y - margin, 0), max(coarse_x - margin, 0)
    fine = None
    for step in FINE_SCALE_STEPS:
        scale = coarse_scale * (1 + step)
        template = scaled_pattern(pattern, scale)
        bottom = min(top + template.shape[0] + 2 * margin, gray.shape[0])
        right = min(left + template.shape[1] + 2 * margin, gray.shape[1])
        window = gray[top:bottom, left:right]
        if window.shape[0] < template.shape[0] or window.shape[1] < template.shape[1]:
            continue
        scores = cv2.matchTemplate(window, template, cv2.TM_CCOEFF_NORMED)
        _, peak, _, (x, y) = cv2.minMaxLoc(scores)
        if fine is None or peak > fine[0]:
            fine = (peak, scale, scores, x, y)
    _, scale, scores, x, y = fine
    if abs(scale - 1) < 1e-9:
        # Same size as the model: whole pixels, so the art is not blurred by interpolation.
        return match_score, scale, left + x - pad, top + y
    x, y = subpixel(scores, x, y)
    return match_score, scale, left + x - pad, top + y


def restore(img, a, b, scale, x, y):
    height, width = img.shape[:2]
    transform = placement(scale, x, y)
    # Outside the model area a = 1 and b = 0, which leaves pixels unchanged.
    a_img = cv2.warpAffine(a, transform, (width, height), flags=cv2.INTER_LINEAR, borderValue=(1, 1, 1))
    b_img = cv2.warpAffine(b, transform, (width, height), flags=cv2.INTER_LINEAR, borderValue=(0, 0, 0))
    fixed = (img.astype(np.float32) - b_img) / np.maximum(a_img, MIN_A)
    return np.round(np.clip(fixed, 0, 255)).astype(np.uint8)


def smooth_letter_edges(img, a, scale, x, y):
    """Blends the letter edges of a restored image into the colors just beside them.

    Each post renders the sharp letter edges slightly differently,
    so restoring with the shared model leaves a thin outline (a bit darker or lighter than its surroundings),
    visible only where the art is smooth.
    Only the edge pixels change, and only where the art beside them is smooth.
    """
    height, width = img.shape[:2]
    opacity = cv2.warpAffine(a.mean(2), placement(scale, x, y), (width, height), borderValue=1)
    gradient = np.hypot(cv2.Sobel(opacity, cv2.CV_32F, 1, 0), cv2.Sobel(opacity, cv2.CV_32F, 0, 1)) / 8
    edge = cv2.dilate((gradient > EDGE_GRADIENT / scale).astype(np.uint8), np.ones((3, 3), np.uint8))
    beside = (1 - edge).astype(np.float32)
    pixels = img.astype(np.float32)
    # Blur without the edge pixels, so the fill only uses colors from beside the edge.
    fill = cv2.GaussianBlur(pixels * beside[..., None], (0, 0), EDGE_FILL_SIGMA)
    fill /= np.maximum(cv2.GaussianBlur(beside, (0, 0), EDGE_FILL_SIGMA), 1e-3)[..., None]
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY).astype(np.float32)
    detail = np.abs(gray - cv2.GaussianBlur(gray, (0, 0), 2)) * beside
    art_detail = cv2.GaussianBlur(detail, (0, 0), 4) / np.maximum(cv2.GaussianBlur(beside, (0, 0), 4), 1e-3)
    smooth = np.clip((BUSY_ART - art_detail) / (BUSY_ART - SMOOTH_ART), 0, 1)
    difference = np.abs(fill - pixels).max(2)
    faint = np.clip((ART_LINE_MIN - difference) / (ART_LINE_MIN - OUTLINE_MAX), 0, 1)
    weight = (cv2.GaussianBlur(edge.astype(np.float32), (0, 0), 0.7) * smooth * faint)[..., None]
    return np.round(np.clip(pixels * (1 - weight) + fill * weight, 0, 255)).astype(np.uint8)


_model = None


def process(src, dst):
    """Returns "restored ..." with where the overlay was found, or the reason the image was copied unchanged."""
    global _model
    if _model is None:
        _model = load_model()
    a, b, pattern = _model
    # An unchanged copy keeps the source format, so it keeps the source extension too.
    copy_dst = os.path.splitext(dst)[0] + os.path.splitext(src)[1]
    img = cv2.imread(src, cv2.IMREAD_COLOR)
    if img is None:
        shutil.copyfile(src, copy_dst)
        return "copied unchanged, cannot decode"
    match_score, scale, x, y = locate(img, pattern)
    if match_score < MIN_MATCH_SCORE:
        shutil.copyfile(src, copy_dst)
        return f"copied unchanged, no overlay found ({img.shape[1]}x{img.shape[0]}, best match={match_score:.0f})"
    restored = smooth_letter_edges(restore(img, a, b, scale, x, y), a, scale, x, y)
    if not cv2.imwrite(dst, restored):
        raise OSError(f"cannot write {dst}")
    return f"restored, overlay at x={x:.1f} y={y:.1f} scale={scale:.4f} match={match_score:.0f}"


def process_safe(src, dst):
    try:
        return process(src, dst)
    except OSError as e:
        return f"FAILED, {e}"


def main():
    if not os.path.isdir(INPUT):
        print(f"{os.path.basename(INPUT)}: {process_safe(INPUT, OUTPUT)}")
        return
    os.makedirs(OUTPUT, exist_ok=True)
    names = sorted(n for n in os.listdir(INPUT) if n.lower().endswith((".png", ".jpg", ".jpeg")))
    sources = [os.path.join(INPUT, n) for n in names]
    # Restored images are written as PNG, so they keep the source name with a .png extension.
    targets = [os.path.join(OUTPUT, os.path.splitext(n)[0] + ".png") for n in names]
    restored = failed = 0
    with ProcessPoolExecutor() as pool:
        for i, (name, result) in enumerate(zip(names, pool.map(process_safe, sources, targets)), 1):
            restored += result.startswith("restored")
            failed += result.startswith("FAILED")
            print(f"[{i}/{len(names)}] {result}: {name}", flush=True)
    print(f"done: {restored} restored, {len(names) - restored - failed} copied unchanged, "
          f"{failed} failed, written to {OUTPUT}")


if __name__ == "__main__":
    main()
