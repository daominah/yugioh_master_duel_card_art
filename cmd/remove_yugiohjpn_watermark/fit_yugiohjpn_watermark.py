"""Fits the per-pixel "SAMPLE" overlay model used by remove_yugiohjpn_watermark.py.

Usage: set ART_DIR, CARD_DIR and OUT_MODEL below, then run
    python fit_yugiohjpn_watermark.py

ART_DIR: Instagram images named like {name}_{cardID}_iyugiohjpn.png (any width, only 1080 wide are used,
images without a card id right before "_iyugiohjpn" are skipped).
CARD_DIR: clean card art named {cardID}.png (e.g. the upscaled Master Duel art).

Each Instagram image is aligned to its clean card art (feature matching, then ECC refinement).
In the overlay strip, for every pixel, shown = a * original + b is fitted across all images
(least squares, then pairs with a bad fit are dropped and pixels are reweighted twice).
Re-run it only if the overlay itself changes (letters or opacity); position and scale are found per image.
"""
import os
import re
from concurrent.futures import ProcessPoolExecutor
from functools import partial

import cv2
import numpy as np


HERE = os.path.dirname(os.path.abspath(__file__))

ART_DIR = r"D:\tmp_process_MD_file\yugiohjpn_arts_renamed"
CARD_DIR = r"D:\tmp_process_MD_file\card_id_up2048\upscayl_png_remacri-4x_4x"
OUT_MODEL = os.path.join(HERE, "watermark_model.npz")
IMAGE_WIDTH = 1080
STRIP_TOP, STRIP_BOTTOM = 360, 40
STRIP_HEIGHT = STRIP_TOP - STRIP_BOTTOM


def align(art_dir, card_dir, name):
    """Returns (name, (shown strip, clean strip, valid mask)) or (name, None)."""
    card_id = re.search(r"_(\d+)_iyugiohjpn", name).group(1)
    shown = cv2.imread(os.path.join(art_dir, name))
    card = cv2.imread(os.path.join(card_dir, card_id + ".png"))
    if shown is None or card is None or shown.shape[1] != IMAGE_WIDTH:
        return name, None
    h, w = shown.shape[:2]
    card = cv2.resize(card, (w, w), interpolation=cv2.INTER_AREA)
    g_shown = cv2.cvtColor(shown, cv2.COLOR_BGR2GRAY)
    g_card = cv2.cvtColor(card, cv2.COLOR_BGR2GRAY)
    sift = cv2.SIFT_create(5000)
    k1, d1 = sift.detectAndCompute(g_shown, None)
    k2, d2 = sift.detectAndCompute(g_card, None)
    matches = cv2.BFMatcher().knnMatch(d2, d1, k=2)
    good = [m[0] for m in matches if len(m) == 2 and m[0].distance < 0.8 * m[1].distance]
    if len(good) < 8:
        return name, None
    src = np.float32([k2[g.queryIdx].pt for g in good])
    dst = np.float32([k1[g.trainIdx].pt for g in good])
    M, _ = cv2.estimateAffinePartial2D(src, dst, method=cv2.RANSAC, ransacReprojThreshold=3)
    if M is None:
        return name, None
    warped = cv2.warpAffine(card, M, (w, h), flags=cv2.INTER_CUBIC)
    coverage = cv2.warpAffine(np.full((w, w), 255, np.uint8), M, (w, h))
    valid = cv2.erode((coverage > 250).astype(np.uint8), np.ones((11, 11)))
    # Refine the alignment on rows above the overlay, where both images show the same art.
    mask = valid.copy()
    mask[h - STRIP_TOP - 20:] = 0
    try:
        warp = np.eye(2, 3, dtype=np.float32)
        criteria = (cv2.TERM_CRITERIA_EPS | cv2.TERM_CRITERIA_COUNT, 80, 1e-6)
        _, warp = cv2.findTransformECC(
            g_shown.astype(np.float32), cv2.cvtColor(warped, cv2.COLOR_BGR2GRAY).astype(np.float32),
            warp, cv2.MOTION_AFFINE, criteria, mask, 7)
        flags = cv2.INTER_CUBIC + cv2.WARP_INVERSE_MAP
        warped = cv2.warpAffine(warped, warp, (w, h), flags=flags)
        coverage2 = cv2.warpAffine(valid * 255, warp, (w, h), flags=cv2.INTER_LINEAR + cv2.WARP_INVERSE_MAP)
        valid = cv2.erode((coverage2 > 250).astype(np.uint8), np.ones((5, 5)))
    except cv2.error:
        pass
    rows = slice(h - STRIP_TOP, h - STRIP_BOTTOM)
    return name, (shown[rows], warped[rows], valid[rows] > 0)


def fit(samples, weights=None):
    shape = (STRIP_HEIGHT, IMAGE_WIDTH, 3)
    n, sx, sy, sxx, sxy = (np.zeros(shape, np.float32) for _ in range(5))
    for i, (shown, clean, valid) in enumerate(samples):
        m = valid[..., None].astype(np.float32)
        if weights is not None:
            m = m * weights[i]
        n += m
        sx += clean * m
        sy += shown * m
        sxx += clean * clean * m
        sxy += clean * shown * m
    n = np.maximum(n, 1e-3)
    mean_x, mean_y = sx / n, sy / n
    a = (sxy / n - mean_x * mean_y) / np.maximum(sxx / n - mean_x * mean_x, 1.0)
    b = mean_y - a * mean_x
    return np.clip(a, 0.05, 1.2), b


def blur(x):
    return cv2.GaussianBlur(x, (0, 0), 1.5)


def residual(a, b, shown, clean):
    restored = np.clip((shown - b) / a, 0, 255)
    return np.abs(blur(restored) - blur(clean)).mean(2)


def robust_fit(samples):
    a, b = fit(samples)
    errors = [residual(a, b, s, c)[v].mean() for s, c, v in samples]
    kept = [s for s, e in zip(samples, errors) if e < np.median(errors) * 2.2]
    a, b = fit(kept)
    for _ in range(2):
        weights = [(1 / (1 + (residual(a, b, s, c) / 6) ** 2))[..., None].astype(np.float32) for s, c, _ in kept]
        a, b = fit(kept, weights)
    return a, b, len(kept)


def main():
    art_dir, card_dir, out_path = ART_DIR, CARD_DIR, OUT_MODEL
    # Only the card art itself pairs with clean card art:
    # group art has no card id, and variants put a word between the id and "_iyugiohjpn" (like _concept, _chibi).
    names = sorted(n for n in os.listdir(art_dir) if n.endswith(".png") and re.search(r"_\d+_iyugiohjpn", n))
    with ProcessPoolExecutor(8) as pool:
        samples = [(s.astype(np.float32), c.astype(np.float32), v)
                   for _, r in pool.map(partial(align, art_dir, card_dir), names, chunksize=2) if r for s, c, v in [r]]
    print(f"aligned {len(samples)} of {len(names)} images")
    a, b, kept = robust_fit(samples)
    np.savez_compressed(out_path, a=a.astype(np.float16), b=b.astype(np.float16))
    print(f"model from {kept} images written to {out_path}")


if __name__ == "__main__":
    main()
