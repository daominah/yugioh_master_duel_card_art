# remove_yugiohjpn_watermark

Removes the semi-transparent "SAMPLE" overlay from card art posted on the Instagram page
<https://www.instagram.com/yugioh_cardgame_official_jpn>

## Run

```bash
pip install -r cmd/remove_yugiohjpn_watermark/requirements.txt
python cmd/remove_yugiohjpn_watermark/remove_yugiohjpn_watermark.py
```

Set `INPUT` and `OUTPUT` at the top of the script (no command line arguments).
They are both files or both directories.
Images without the overlay are copied unchanged.
Each image gets one log line: where the overlay was found, or why it was skipped.

## Files

- `remove_yugiohjpn_watermark.py`: restores an image using the model.
- `fit_yugiohjpn_watermark.py`: builds the model.
  Needed only if the overlay itself changes, not for normal use.
- `watermark_model.npz`: the model, committed so nobody has to refit.
- `requirements.txt`: Python dependencies.
- `resize_short_side_2048.sh`: after upscaling the restored images (Upscayl),
  shrinks them so the shorter side is 2048 pixels (ImageMagick, Lanczos in linear light).
  Images already at 2048 or smaller are copied unchanged.
  An odd width or height is cropped by one pixel to make it even.

## The model

There is no machine learning and nothing to download.
The overlay is a fixed layer, so every pixel of the covered strip follows
`shown = a * original + b`, where `a` is how much of the art shows through
and `b` is the colour the overlay adds.
`watermark_model.npz` stores `a` and `b`
(float16, shape 320x1080x3, about 3 MB) for the strip of rows `H-360` to `H-40`
in an image that is 1080 pixels wide.
Restoring is `(shown - b) / a`, with `a` floored at 0.2 to avoid amplifying noise.

The bottom rows of the strip hold the copyright footer layer
(a gray band and the top of the footer text).
The fit treats them as overlay because the clean art has no footer,
but restoring them smears the footer text.
The footer is not the watermark, so the script fades the model to "no change"
below the letters (model row 304) and the footer stays as posted.

### Where the model came from

`fit_yugiohjpn_watermark.py` (set `ART_DIR`, `CARD_DIR`, `OUT_MODEL` at its top) fits it from pairs of
the same card:

- Instagram image, with the overlay (123 images renamed as
  `{name}_{cardID}_iyugiohjpn.png`, matched to card ids by image similarity).
- Clean card art, the Master Duel art upscaled to 2048 pixels
  (`card_id_up2048`, from Upscayl with the `remacri-4x` model).

Steps:

- Align each clean art to its Instagram image with feature matching,
  then refine with intensity based alignment.
  107 of 123 pairs aligned.
- Per pixel, fit `a` and `b` by least squares across all aligned pairs.
- Drop pairs that fit badly (alignment failures or art that differs),
  then refit with outliers down-weighted.
  90 pairs were kept.

Accuracy on pairs held out of the fit:
the average error against the clean art drops from about 18 to about 3 (scale 0 to 255).
A faint outline of the letter edges can remain in light, low-contrast areas,
see [Smoothing the letter edges](#smoothing-the-letter-edges).

## Finding the overlay

The overlay is not at the same place on every post,
so the script searches each image before restoring:

- Pattern: what the overlay does to mid-gray art, letters only
  (the footer is also on posts without the overlay, so it would cause false matches).
- Search: slide the pattern over the edges of the image (`cv2.matchTemplate`)
  at scales within 3% of `image_width / 1080`.
  If nothing is found, retry over a wide range of scales (0.5 to 1.0 of that).
- Decision: the match score is how far the best match stands out
  from the rest of the image, in standard deviations.
  The overlay counts as found at 16 or more.
- Refine: finer scale steps and sub-pixel position around the best match.
- Restore: resize and move the model onto the image,
  so the art itself is never resampled.

Results on the 308 images saved from the page:

- 252 restored, 56 copied unchanged.
- Images with the overlay scored 19 or more, images without it 13 or less.
  The 56 skipped were checked by eye: sketches, photos, banners, logos and avatars.
- Positions found besides the standard bottom spot:
  1080x1350 portraits shifted 6 to 31 pixels from the 1080x1080 spot,
  4 posts with the overlay near the top,
  posts resized to 1024 to 1081 or 2047 to 2048 pixels wide,
  and 1024x1024 posts with black side bars, where the overlay is about 0.65 of the usual size.

Accuracy on the 116 restored images that have clean card art,
measured over the letters (scale 0 to 255):

| Version | Average error |
|---|---|
| No restore | 21.05 |
| Fixed bottom spot, no search | 2.83 |
| With search | 2.47 |

No image got worse with the search; 10 improved by more than 0.5.

## Smoothing the letter edges

Each post renders the sharp letter edges slightly differently,
so restoring with the shared model leaves a thin outline of "SAMPLE":
a line a bit darker or lighter than the art on both sides.
It is only visible where the art is smooth, for example the light yellow in `the_iris_swordsoul`.

After restoring, the script blends each letter edge pixel (about 3 pixels wide)
into the colors just beside it, but only where both hold:

- The art beside the edge is smooth (busy art hides the outline anyway).
- The pixel differs from its surroundings by at most a few gray levels
  (a larger difference is more likely an art line crossing the letter edge, so it is kept).

Checked by eye on iris, busy art and art with thin lines crossing the letters:
the outline is gone where it was visible, and art lines are kept.
Against the clean card art, the error on letter edges over smooth art drops from 2.68 to 2.47.
The clean card art is a separate upscaled render,
so it cannot tell a removed outline from softened art detail;
judge this step by eye.

## Known limits

- A faint outline of the letter edges can remain over moderately detailed art,
  where the smoothing backs off to protect the art.
- Only 1080 pixel wide posts have enough pairs to fit a model (about 20 or more needed).
  Other sizes reuse the 1080 model resized, which is slightly less precise on letter edges.
