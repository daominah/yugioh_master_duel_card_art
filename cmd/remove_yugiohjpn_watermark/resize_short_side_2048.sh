#!/bin/bash
# Resizes upscaled images so their shorter side is 2048 pixels, keeping the aspect ratio
# (2160x2700 becomes 2048x2560). Written to a new directory, the source is left untouched.
#
# Only shrinks: images with a shorter side of 2048 or less are copied byte for byte,
# because enlarging only blurs and resizing to the same size could shift a few pixel values.
# Resizing in linear light (colorspace RGB) keeps thin bright lines from darkening,
# Lanczos keeps detail when shrinking.
# Odd dimensions are then cropped to even ones.

IMAGEMAGICK="/c/Program Files/ImageMagick-6.9.13-Q16-HDRI"
SRC="/d/tmp_process_MD_file/yugiohjpn_arts_clean/upscayl_png_remacri-4x_2x"
DST="${SRC}_short2048"
SHORT_SIDE=2048

mkdir -p "$DST"
for f in "$SRC"/*.png; do
    short=$("$IMAGEMAGICK/identify.exe" -format "%[fx:min(w,h)]" "$f")
    if [ "$short" -le "$SHORT_SIDE" ]; then
        cp "$f" "$DST/"
        echo "copied ($short): $(basename "$f")"
    else
        "$IMAGEMAGICK/convert.exe" "$f" -colorspace RGB -filter Lanczos \
            -resize "${SHORT_SIDE}x${SHORT_SIDE}^" -colorspace sRGB "$DST/$(basename "$f")"
        echo "resized ($short): $(basename "$f")"
    fi
done

# Crops odd dimensions to even by dropping the last column or row (3641x2048 becomes 3640x2048).
for f in "$DST"/*.png; do
    read -r width height <<< "$("$IMAGEMAGICK/identify.exe" -format "%w %h" "$f")"
    if [ $((width % 2)) -eq 1 ] || [ $((height % 2)) -eq 1 ]; then
        even_width=$((width / 2 * 2))
        even_height=$((height / 2 * 2))
        "$IMAGEMAGICK/convert.exe" "$f" -crop "${even_width}x${even_height}+0+0" +repage "$f"
        echo "cropped ${width}x${height} to ${even_width}x${even_height}: $(basename "$f")"
    fi
done
