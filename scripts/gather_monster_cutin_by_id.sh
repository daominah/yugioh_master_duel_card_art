#!/bin/bash
#
# gather_monster_cutin_by_id.sh
#
# Collects monster cut-in (Spine 2D) assets into a flat, editor-ready folder
# (one per card) so the Official Spine editor can import them directly.
# Gathers one card by id (--card), or every card that ships a cut-in (--all).
#
# Only some cards ship a cut-in.
# If the card id has none, the script reports "not found" and exits with code 1.
# To check beforehand, see `monster_cutin.md` at the repo root,
# a table of every card id with a cut-in, its name, and ocg/tcg availability.
#
# A card can exist under both the ocg and tcg region folders, but the two
# are not the same asset (atlas matches, texture pages differ). Pick the region
# with --region ocg|tcg|both. Default (no --region) prefers ocg, falling back to tcg
# if the card has no ocg cut-in.
#
# Directory layout under the cut-in root (input_root), one tier shown;
# sd is the same layout, smaller scale and fewer/smaller texture pages:
#   p<id>/
#     highend_hd/
#       P<id>JS.asset                          # skeleton, or P<id>JS.json (see note)
#       Skeleton Prefab Mesh [SpineP<id>].obj  # static bind-pose mesh, not gathered
#       <scale>/                               # e.g. 0.56, varies per card
#         <base>.atlas                         # base varies, see note below
#         <base>.png, <base>_2.png, ...
#     sd/
#       ... same layout ...
#
# What it gathers, for each picked region and tier (highend_hd, sd) that exist:
#   - the skeleton (P<id>JS.json or P<id>JS.asset, both are raw Spine JSON; the
#     JS casing varies, e.g. P<id>Js.json, so the match is case-insensitive),
#     copied as <out_base>.json, where out_base is the per-card output folder's
#     name (name_prefix + p<id>, e.g. centur_ion_legatia_p19375), so its base
#     name matches both the atlas, for auto-pairing, and the folder.
#     The .json versus .asset extension does not track the tier reliably:
#     some cards use .asset in both tiers, some use .json in both, some split
#     by tier. It depends on the card and region.
#   - the atlas and its texture pages, nested in a scale-named subfolder (e.g. 0.56).
#     The atlas base name is inconsistent: usually P<id>, but some cards use a
#     generic ForUnity/Forunity, and a few use a transposed id (P9696 for card
#     6969). The script matches the atlas by extension, copies everything renamed
#     to <out_base>, and rewrites the atlas's page references to match, so the
#     output is uniform regardless of the source name. A stray source atlas at the
#     tier root (not in the scale subfolder) is ignored.
# The static bind-pose mesh (.obj) is not gathered: the Spine editor does not use it.
#
# Output layout depends on how many region/tier combinations are gathered:
#   - exactly one (the common case): flat at <output_dir>/<name>_p<id>/.
#   - more than one (--region both and/or --tier both): one subfolder per combination,
#     since their file names collide (each holds a <name>_p<id>.atlas, .png, ...).
#
# Flags:
#   --card <card_id>    card id to gather (required unless --all).
#   --all               gather every card listed in monster_cutin.md.
#   --input <path>      cut-in root to read from (default: the extracted monstercutin dir).
#   --output <path>     where to write the gathered folder (default: a tmp view dir).
#   --tier <tier>       highend_hd, sd, or both. Default auto-picks highend_hd, else sd.
#   --region <region>   ocg, tcg, or both. Default auto-picks ocg, else tcg.
#   --force             re-gather even if the card's output folder already exists.
#                       Without it, an existing output folder is skipped (resumable).
#   --help              print usage and exit.
#
# Usage:
#   ./scripts/gather_monster_cutin_by_id.sh --card 19375
#   ./scripts/gather_monster_cutin_by_id.sh --all
#   ./scripts/gather_monster_cutin_by_id.sh --all --force
#   ./scripts/gather_monster_cutin_by_id.sh --card 19375 --region both --tier both
#   ./scripts/gather_monster_cutin_by_id.sh --card 19375 --input /d/tmp_process_MD_file_by_path/... --output /d/out

set -euo pipefail

input_root='/d/tmp_process_MD_file_by_path/assets/resourcesassetbundle/duel/timeline/duel/monstercutin'
output_dir='/d/tmp_process_MD_view_monster_cutin'
# Which tier to gather. Default ('') auto-picks highend_hd, the best quality and
# the one you view, falling back to sd (a smaller, lower-detail copy) if the card
# ships no highend_hd cut-in. Pass highend_hd or sd to force one, or both for the two.
tier=''
# Which region to gather. Default ('') auto-picks ocg, falling back to tcg
# if the card has no ocg cut-in. Pass tcg to force tcg, or both for the two.
region=''
card_id=''
# Gather every card in monster_cutin.md instead of a single --card.
gather_all=false
# Re-gather even when the card's output folder already exists.
# Default (false) skips such cards, so an interrupted --all run resumes cheaply.
force=false

usage() {
    echo "Usage: $0 (--card <card_id> | --all) [--input <input_root>] [--output <output_dir>] [--tier highend_hd|sd|both] [--region ocg|tcg|both] [--force]" >&2
    exit 2
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --card)   card_id="$2";    shift 2 ;;
        --all)    gather_all=true; shift ;;
        --input)  input_root="$2"; shift 2 ;;
        --output) output_dir="$2"; shift 2 ;;
        --tier)   tier="$2";       shift 2 ;;
        --region) region="$2";     shift 2 ;;
        --force)  force=true;      shift ;;
        --help)   usage ;;
        *) echo "unknown option: $1" >&2; usage ;;
    esac
done

# Require exactly one of --card or --all.
if [[ "$gather_all" == true && -n "$card_id" ]]; then
    echo "use either --card or --all, not both" >&2
    usage
fi
if [[ "$gather_all" != true && -z "$card_id" ]]; then
    usage
fi
case "$tier" in
    ''|highend_hd|sd|both) ;;
    *) echo "invalid --tier value: $tier (expected highend_hd, sd, or both)" >&2; exit 2 ;;
esac
# Human-readable tier for status lines (the auto default has no single name).
tier_desc="${tier:-highend_hd/sd}"
case "$region" in
    ''|ocg|tcg|both) ;;
    *) echo "invalid --region value: $region (expected ocg, tcg, or both)" >&2; exit 2 ;;
esac

if [[ ! -d "$input_root" ]]; then
    echo "cut-in root not found: $input_root  (edit the --input path)" >&2
    exit 1
fi

# Card name lookup: monster_cutin.md (repo root, sibling of scripts/) maps each
# card id to its name. The name becomes a slug prefix on the output folder
# (e.g. centur_ion_legatia_p19375), and the table also drives --all.
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
monster_cutin_md="$script_dir/../monster_cutin.md"

# Set the card_name and name_prefix globals for one card id.
# Both fall back to empty (bare p<id> folder) if the table or the row is missing.
lookup_card_name() {
    local id="$1"
    card_name=''
    name_prefix=''
    [[ -f "$monster_cutin_md" ]] || return 0
    card_name=$(awk -F'|' -v id="$id" '
        {
            cid = $2; gsub(/^[ \t]+|[ \t]+$/, "", cid)
            if (cid == id) { name = $3; gsub(/^[ \t]+|[ \t]+$/, "", name); print name; exit }
        }
    ' "$monster_cutin_md")
    card_name="${card_name% (alt art)}"
    if [[ -n "$card_name" ]]; then
        name_prefix="$(printf '%s' "$card_name" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9_]/_/g')_"
    fi
}

# Every numeric card id listed in monster_cutin.md, one per line (drives --all).
all_card_ids() {
    awk -F'|' '{ cid = $2; gsub(/^[ \t]+|[ \t]+$/, "", cid); if (cid ~ /^[0-9]+$/) print cid }' "$monster_cutin_md"
}

# Output folder for one region+tier combination.
resolve_out_dir() {
    local r="$1" t="$2" card_out_dir="$output_dir/${name_prefix}p${card_id}"
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

# Whether one region+tier actually ships a usable cut-in (atlas + skeleton present)
# for the current card. Used to auto-pick the tier without copying anything.
tier_available() {
    local r="$1" t="$2" tier_dir="$input_root/$r/p${card_id}/$t"
    [[ -d "$tier_dir" ]] || return 1
    local atlas skeleton
    # The atlas + pages sit in a scale-named subfolder (depth >= 2). Its base name
    # varies (P<id>, ForUnity, a transposed id), so match by extension. Ignore any
    # stray source atlas at the tier root (depth 1).
    atlas=$(find "$tier_dir" -mindepth 2 -type f -name "*.atlas" -print -quit)
    skeleton=$(find "$tier_dir" -maxdepth 1 -type f -iname "P${card_id}JS.*" -print -quit)
    [[ -n "$atlas" && -n "$skeleton" ]]
}

# Copy one region+tier's skeleton, atlas, and pages into a flat output folder.
# Echoes a status line and returns 0 if it gathered files, 1 if absent.
copy_tier() {
    local r="$1" tier_name="$2" out_dir="$3"
    local tier_dir="$input_root/$r/p${card_id}/$tier_name"
    [[ -d "$tier_dir" ]] || return 1

    # Atlas + pages sit in a scale-named subfolder; the atlas base name varies, so
    # match by extension (see tier_available). Skeleton sits at the tier root.
    local atlas skeleton
    atlas=$(find "$tier_dir" -mindepth 2 -type f -name "*.atlas" -print -quit)
    skeleton=$(find "$tier_dir" -maxdepth 1 -type f -iname "P${card_id}JS.*" -print -quit)
    if [[ -z "$atlas" || -z "$skeleton" ]]; then
        echo "  warning: ${r}/${tier_name}: missing atlas or skeleton, skipping." >&2
        return 1
    fi

    local atlas_dir atlas_base
    atlas_dir=$(dirname "$atlas")
    atlas_base=$(basename "$atlas" .atlas)

    mkdir -p "$out_dir"

    # Normalize every file to the card's output base name (name_prefix + p<id>,
    # the same name as the per-card output folder, e.g. centur_ion_legatia_p19375)
    # so the editor auto-pairs them and the output matches standard cards
    # regardless of the source name: rename the skeleton to <base>.json, rewrite
    # the atlas's page-image lines (those ending in .png) from <atlas_base>* to
    # <base>*, and rename the texture pages to match.
    local out_base="${name_prefix}p${card_id}"
    cp "$skeleton" "$out_dir/${out_base}.json"
    sed "/\.png\$/ s/^${atlas_base}/${out_base}/" "$atlas" > "$out_dir/${out_base}.atlas"
    local png fname
    for png in "$atlas_dir/${atlas_base}"*.png; do
        [[ -e "$png" ]] || continue
        fname=$(basename "$png")
        cp "$png" "$out_dir/${out_base}${fname#"$atlas_base"}"
    done

    local page_count
    page_count=$(find "$out_dir" -maxdepth 1 -type f -name "${out_base}*.png" | wc -l)
    echo "  ${r}/${tier_name} -> ${out_dir}  (${page_count} page(s))"
    return 0
}

# Gather one card. Sets the per-card globals (card_id, regions, tiers, combo_count,
# card_name, name_prefix) that the helpers above read. Returns:
#   0  gathered files
#   1  card has no matching cut-in
#   2  skipped, output folder already exists (and --force not given)
gather_card() {
    card_id="$1"

    lookup_card_name "$card_id"

    # Skip cards already gathered, so an interrupted --all run resumes cheaply.
    local card_out_dir="$output_dir/${name_prefix}p${card_id}"
    if [[ "$force" != true && -d "$card_out_dir" ]]; then
        echo "Skipping card $card_id${card_name:+ ($card_name)}: output exists ($card_out_dir); use --force to re-gather"
        return 2
    fi

    # Candidate tiers in preference order, used both to pick a usable region and
    # to settle the tier below. Auto mode (tier unset) considers highend_hd then sd.
    local t r cand_tiers
    case "$tier" in
        both)          cand_tiers=(highend_hd sd) ;;
        highend_hd|sd) cand_tiers=("$tier") ;;
        '')            cand_tiers=(highend_hd sd) ;;
    esac

    # Resolve which region(s) to gather. Auto mode (region unset) prefers ocg,
    # then tcg, settling on the first region that ships a usable cut-in (atlas +
    # skeleton). A region folder can exist yet hold no atlas, so directory
    # presence alone is not enough.
    regions=()
    case "$region" in
        both) regions=(ocg tcg) ;;
        ocg|tcg) regions=("$region") ;;
        '')
            for r in ocg tcg; do
                for t in "${cand_tiers[@]}"; do
                    if tier_available "$r" "$t"; then regions=("$r"); break 2; fi
                done
            done
            # No usable cut-in anywhere: fall back to directory presence so the
            # warning below still names a concrete region.
            if [[ ${#regions[@]} -eq 0 ]]; then
                if [[ -d "$input_root/ocg/p${card_id}" ]]; then
                    regions=(ocg)
                elif [[ -d "$input_root/tcg/p${card_id}" ]]; then
                    regions=(tcg)
                fi
            fi
            ;;
    esac
    if [[ ${#regions[@]} -eq 0 ]]; then
        echo "warning: card $card_id has no cut-in under $input_root" >&2
        return 1
    fi

    # Resolve which tier(s) to gather within the chosen region(s). Auto mode
    # prefers highend_hd, falling back to sd when no highend_hd cut-in exists.
    case "$tier" in
        both) tiers=(highend_hd sd) ;;
        highend_hd|sd) tiers=("$tier") ;;
        '')
            tiers=(highend_hd)  # default if none usable, so copy_tier warns
            for t in highend_hd sd; do
                for r in "${regions[@]}"; do
                    if tier_available "$r" "$t"; then tiers=("$t"); break 2; fi
                done
            done
            ;;
    esac

    # Combination count decides the output layout (see header note).
    combo_count=$((${#regions[@]} * ${#tiers[@]}))

    echo "Gathering cut-in for card $card_id${card_name:+ ($card_name)} (region: ${regions[*]}, tier: ${tiers[*]})"
    local gathered=false out_dir
    for r in "${regions[@]}"; do
        for t in "${tiers[@]}"; do
            out_dir=$(resolve_out_dir "$r" "$t")
            if copy_tier "$r" "$t" "$out_dir"; then
                gathered=true
            fi
        done
    done
    if [[ "$gathered" != true ]]; then
        echo "warning: card $card_id has no ${tier_desc} cut-in under ${regions[*]}" >&2
        return 1
    fi
    return 0
}

# Dispatch: every card in the table (--all), or a single --card.
if [[ "$gather_all" == true ]]; then
    if [[ ! -f "$monster_cutin_md" ]]; then
        echo "table not found: $monster_cutin_md  (needed for --all)" >&2
        exit 1
    fi

    # Card ids already gathered, read once via a single `find` pass over
    # $output_dir before the loop. Without this, skipping an already-gathered
    # card would still cost a lookup_card_name awk fork plus a directory stat
    # per id, every run, even though the answer never changes during the run.
    # Every output folder this script creates ends in p<id>, so the trailing
    # digits after the last "p" recover the id regardless of name_prefix.
    declare -A existing_output_ids=()
    if [[ "$force" != true ]]; then
        while IFS= read -r id; do
            [[ -n "$id" ]] && existing_output_ids["$id"]=1
        done < <(find "$output_dir" -mindepth 1 -maxdepth 1 -type d -name '*p[0-9]*' -print 2>/dev/null |
            sed -E 's#.*/##; s/^.*p([0-9]+)$/\1/')
    fi

    ok=0
    existing=0
    missing=0
    while IFS= read -r id; do
        if [[ "$force" != true && -n "${existing_output_ids[$id]:-}" ]]; then
            echo "Skipping card $id: output exists; use --force to re-gather"
            existing=$((existing + 1))
            echo
            continue
        fi
        rc=0
        gather_card "$id" || rc=$?
        case $rc in
            0) ok=$((ok + 1)) ;;
            2) existing=$((existing + 1)) ;;
            *) missing=$((missing + 1)) ;;
        esac
        echo
    done < <(all_card_ids)
    echo "Done. Gathered $ok card(s) into $output_dir" \
         "($existing already present, $missing with no ${tier_desc} cut-in skipped)."
else
    rc=0
    gather_card "$card_id" || rc=$?
    case $rc in
        0)
            echo
            echo "Done. Open the ${name_prefix}p${card_id}.json in the Official Spine editor (File -> Import Data)."
            ;;
        2) ;;  # already-present message already printed; nothing to gather
        *) exit 1 ;;
    esac
fi
