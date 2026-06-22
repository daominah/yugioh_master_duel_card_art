#!/bin/bash
#
# gather_monster_cutin_by_id.sh
#
# Collects one card's monster cut-in (Spine 2D) assets into a flat, editor-ready folder,
# so the Official Spine editor can import them directly.
#
# Only some cards ship a cut-in.
# If the card id has none, the script reports "not found" and exits with code 1.
# To check beforehand, see `monster_cutin.md` at the repo root,
# a table of every card id with a cut-in, its name, and ocg/tcg availability.
#
# A card can exist under both the ocg and tcg region folders, but the two
# are not the same asset (atlas matches, texture pages differ). Pick the region
# with -r ocg|tcg|all. Default (no -r) prefers ocg, falling back to tcg
# if the card has no ocg cut-in.
#
# Directory layout under the cut-in root (input_root), one tier shown;
# sd is the same layout, smaller scale and fewer/smaller texture pages:
#   p<id>/
#     highend_hd/
#       P<id>JS.asset                          # skeleton, or P<id>JS.json (see note)
#       Skeleton Prefab Mesh [SpineP<id>].obj  # static bind-pose mesh, not gathered
#       <scale>/                               # e.g. 0.56, varies per card
#         P<id>.atlas
#         P<id>.png, P<id>_2.png, ...
#     sd/
#       ... same layout ...
#
# What it gathers, for each picked region and tier (highend_hd, sd) that exist:
#   - the skeleton (P<id>JS.json or P<id>JS.asset, both are raw Spine JSON),
#     copied as P<id>.json so its base name matches the atlas for auto-pairing.
#     The .json versus .asset extension does not track the tier reliably:
#     some cards use .asset in both tiers, some use .json in both, some split
#     by tier. It depends on the card and region.
#   - the atlas (P<id>.atlas) and its texture pages (P<id>.png, P<id>_2.png, ...),
#     which the extraction nests in a scale-named subfolder (e.g. 0.56).
# The static bind-pose mesh (.obj) is not gathered: the Spine editor does not use it.
#
# Output layout depends on how many region/tier combinations are gathered:
#   - exactly one (the common case): flat at <output_dir>/p<id>/.
#   - more than one (-r all and/or -t all): one subfolder per combination,
#     since their file names collide (each holds a P<id>.atlas, P<id>.png, ...).
#
# Usage:
#   ./scripts/gather_monster_cutin_by_id.sh -c 19375
#   ./scripts/gather_monster_cutin_by_id.sh -c 19375 -r all -t all
#   ./scripts/gather_monster_cutin_by_id.sh -c 19375 -i /d/tmp_process_MD_file_by_path/... -o /d/out

set -euo pipefail

input_root='/d/tmp_process_MD_file_by_path/assets/resourcesassetbundle/duel/timeline/duel/monstercutin'
output_dir='/d/tmp_process_MD_view_monster_cutin'
# Which tiers to gather. Default is highend_hd, the best quality and the one you view.
# sd is a smaller, lower-detail copy; pass sd or all only if needed.
tier='highend_hd'
# Which region to gather. Default ('') auto-picks ocg, falling back to tcg
# if the card has no ocg cut-in. Pass tcg to force tcg, or all for both.
region=''
card_id=''

usage() {
    echo "Usage: $0 -c <card_id> [-i <input_root>] [-o <output_dir>] [-t highend_hd|sd|all] [-r ocg|tcg|all]" >&2
    exit 2
}

while getopts ':c:i:o:t:r:h' opt; do
    case "$opt" in
        c) card_id="$OPTARG" ;;
        i) input_root="$OPTARG" ;;
        o) output_dir="$OPTARG" ;;
        t) tier="$OPTARG" ;;
        r) region="$OPTARG" ;;
        h) usage ;;
        *) usage ;;
    esac
done

[[ -z "$card_id" ]] && usage
case "$tier" in
    highend_hd|sd|all) ;;
    *) echo "invalid -t value: $tier (expected highend_hd, sd, or all)" >&2; exit 2 ;;
esac
case "$region" in
    ''|ocg|tcg|all) ;;
    *) echo "invalid -r value: $region (expected ocg, tcg, or all)" >&2; exit 2 ;;
esac

if [[ ! -d "$input_root" ]]; then
    echo "cut-in root not found: $input_root  (edit the -i path)" >&2
    exit 1
fi

if [[ "$tier" == 'all' ]]; then
    tiers=(highend_hd sd)
else
    tiers=("$tier")
fi

# Resolve which region(s) to gather. Auto mode (region unset) tries ocg first,
# then tcg, and settles on whichever one the card actually has.
regions=()
case "$region" in
    all) regions=(ocg tcg) ;;
    ocg|tcg) regions=("$region") ;;
    '')
        if [[ -d "$input_root/ocg/p${card_id}" ]]; then
            regions=(ocg)
        elif [[ -d "$input_root/tcg/p${card_id}" ]]; then
            regions=(tcg)
        fi
        ;;
esac

if [[ ${#regions[@]} -eq 0 ]]; then
    echo "warning: card $card_id has no cut-in under $input_root" >&2
    exit 1
fi

# Combination count decides the output layout (see header note).
combo_count=$((${#regions[@]} * ${#tiers[@]}))

# Look up the card name in monster_cutin.md (repo root, sibling of scripts/)
# and append a standardized slug to the output folder, e.g. p19375_centur_ion_legatia.
# Falls back to the bare p<id> folder if the table or the row is missing.
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
monster_cutin_md="$script_dir/../monster_cutin.md"
name_suffix=''
if [[ -f "$monster_cutin_md" ]]; then
    card_name=$(awk -F'|' -v id="$card_id" '
        {
            cid = $2; gsub(/^[ \t]+|[ \t]+$/, "", cid)
            if (cid == id) { name = $3; gsub(/^[ \t]+|[ \t]+$/, "", name); print name; exit }
        }
    ' "$monster_cutin_md")
    card_name="${card_name% (alt art)}"
    if [[ -n "$card_name" ]]; then
        name_suffix="_$(printf '%s' "$card_name" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9_]/_/g')"
    fi
fi

# Output folder for one region+tier combination.
resolve_out_dir() {
    local r="$1" t="$2" card_out_dir="$output_dir/p${card_id}${name_suffix}"
    if [[ $combo_count -le 1 ]]; then
        echo "$card_out_dir"
    elif [[ ${#regions[@]} -gt 1 && ${#tiers[@]} -eq 1 ]]; then
        echo "$card_out_dir/$r"
    elif [[ ${#regions[@]} -eq 1 && ${#tiers[@]} -gt 1 ]]; then
        echo "$card_out_dir/$t"
    else
        echo "$card_out_dir/${r}_${t}"
    fi
}

# Copy one region+tier's skeleton, atlas, and pages into a flat output folder.
# Echoes a status line and returns 0 if it gathered files, 1 if absent.
copy_tier() {
    local r="$1" tier_name="$2" out_dir="$3"
    local tier_dir="$input_root/$r/p${card_id}/$tier_name"
    [[ -d "$tier_dir" ]] || return 1

    # Atlas and pages sit in a scale-named subfolder; find the atlas anywhere under the tier.
    local atlas
    atlas=$(find "$tier_dir" -type f -name "P${card_id}.atlas" -print -quit)
    # Skeleton sits at the tier root, as .json or .asset (both raw Spine JSON).
    local skeleton
    skeleton=$(find "$tier_dir" -maxdepth 1 -type f -name "P${card_id}JS.*" -print -quit)
    if [[ -z "$atlas" || -z "$skeleton" ]]; then
        echo "  warning: ${r}/${tier_name}: missing atlas or skeleton, skipping." >&2
        return 1
    fi

    mkdir -p "$out_dir"

    # Rename skeleton to match the atlas base name so the editor auto-pairs them.
    cp "$skeleton" "$out_dir/P${card_id}.json"
    cp "$atlas" "$out_dir/"
    local atlas_dir
    atlas_dir=$(dirname "$atlas")
    find "$atlas_dir" -maxdepth 1 -type f -name "P${card_id}*.png" -exec cp {} "$out_dir/" \;

    local page_count
    page_count=$(find "$out_dir" -maxdepth 1 -type f -name "P${card_id}*.png" | wc -l)
    echo "  ${r}/${tier_name} -> ${out_dir}  (${page_count} page(s))"
    return 0
}

echo "Gathering cut-in for card $card_id${card_name:+ ($card_name)} (region: ${regions[*]}, tier: $tier)"
echo

gathered=false
for r in "${regions[@]}"; do
    for t in "${tiers[@]}"; do
        out_dir=$(resolve_out_dir "$r" "$t")
        if copy_tier "$r" "$t" "$out_dir"; then
            gathered=true
        fi
    done
done

if [[ "$gathered" != true ]]; then
    echo "warning: card $card_id has no $tier cut-in under ${regions[*]}" >&2
    exit 1
fi

echo
echo "Done. Open the P${card_id}.json in the Official Spine editor (File -> Import Data)."
