# YuGiOh Master Duel all cards art

Extract card image data from "Yu-Gi-Oh! Master Duel" and
rename files from Konami card IDs to card names.

## Steps

### Download the game from Steam

[steampowered.com/app/YuGiOh_Master_Duel](https://store.steampowered.com/app/1449850/YuGiOh_Master_Duel/)

### Download AssetStudio

AssetStudio is an app to extract Unity game assets.
Download from one of the following URLs:

- [github.com/aelurum/AssetStudio](https://github.com/aelurum/AssetStudio)
- [github.com/Perfare/AssetStudio](https://github.com/Perfare/AssetStudio/releases) (original, archived)

Default .NET version on Windows is v4.x

### Prepare Steam account assets

On my Windows, each Steam accounts have its own game assets, located in
`D:\game\SteamLibrary\steamapps\common\Yu-Gi-Oh!  Master Duel\LocalData`

```bash
cd "/d/game/SteamLibrary/steamapps/common/Yu-Gi-Oh!  Master Duel/LocalData"
ls -lh --time-style=+%Y-%m-%d
# Output:
# 70102374 minahdao, English card art
# 2509bcbc bixuzofa, Japanese card art
# dca6ade4 em_chef_tft, duplicated English card art, move before extract

mv dca6ade4 /d/game/SteamLibrary/steamapps/common/localdata_dca6ade4
mv "/d/game/SteamLibrary/steamapps/common/Yu-Gi-Oh!  Master Duel/LocalSave/dca6ade4" /d/game/SteamLibrary/steamapps/common/localsave_dca6ade4

ls /d/game/SteamLibrary/steamapps/common | grep local
# Output:
# localdata_dca6ade4/
# localsave_dca6ade4/
```

### Extract game assets

#### Load folder

Start `AssetStudio` (probably as administrator, so it has less memory errors).

File: Load folder: `D:\game\SteamLibrary\steamapps\common\Yu-Gi-Oh!  Master Duel`

This takes about 30 minutes and use almost all the computer remaining memory.

#### Export

* Options: Export options: Group exported assets by: container path
* Filter Type: choose all EXCEPT the following types:
  - MonoBehaviour (a lot of human unreadable files)
  - Animator (probably they cause exporting errors)
* Export: Filtered assets: choose output to `D:\tmp_process_MD_file_by_path`
  (make sure the directory exists and empty).

This takes about 3 hours,
occasionally show errors and stuck, require human to click OK.

The result card arts as PNG, named as Konami card ID,
most are in dir `assets/resources/card/images/illust/common`.

#### Export by chunks to avoid out-of-memory (optional)

The GUI "Load folder" approach loads all ~38k bundles (about 13 GB) at once,
decompresses blocks into RAM, then decodes textures to RGBA in parallel.
On a 32 GB machine this can commit ~50 GB and crash with out-of-memory,
which is why some card arts come out missing.

`scripts/extract_md_assets_chunks.ps1` avoids this by driving the CLI build
([AssetStudioModCLI](https://github.com/aelurum/AssetStudio))
once per hex bucket of the `LocalData` store.
Each invocation loads only one bucket (about 150 bundles),
exports, then exits and frees all memory before the next bucket,
so peak RAM stays small.

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\extract_md_assets_chunks.ps1
```

Key points, all set as defaults at the top of the script:

- `-t tex2d,sprite,textAsset,audio,mesh`: every type present in the data
  except MonoBehaviour and Animator (same filter as the GUI steps above).
- `-g container`: keeps the same output tree as a single GUI run,
  because the path comes from each bundle's container metadata, not the bucket.
- `--max-export-tasks 2`: caps concurrent texture decodes to limit peak memory.
  We skip `--decompress-to-disk`: per-bucket loading already bounds memory,
  so keeping decompression in RAM is faster.
- No `-r` flag, so files that already exist are skipped.
  This makes the run resumable, and pointing `OutputDir` at an existing
  extraction fills only the assets a previous crashed run missed.
- `OutputDir` defaults to `D:\tmp_process_MD_file_by_path` (the parent of `assets`),
  not `...\assets`. The CLI container paths already begin with `assets/`,
  so the script writes a single `assets\resources\card\...` tree.
  Pointing it at `...\assets` would nest a redundant `assets\assets\...` level.

The game ships card art in two places, and the script covers both:

- `LocalData\<account>\0000\`: the main store (about 13 GB),
  downloaded and updated over time, processed one hex bucket per run.
- `masterduel_Data\StreamingAssets\AssetBundle\`: a small bootstrap set
  (about 100 MB) shipped with the game for the tutorial/first run.
  It also holds card art, under an `assets/resources/card/...` container,
  and is exported in one final pass after the buckets.

##### Keep different variants across accounts (`-KeepDifferentVariants`)

By default the CLI skips an output file whenever one of the same name exists,
regardless of content.
That hides the case where a second account has different art for the same card id,
for example a censored versus an uncensored illustration.

Run with `-KeepDifferentVariants` to compare by content instead:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\extract_md_assets_chunks.ps1 `
  -InputRoot 'D:\game\SteamLibrary\steamapps\common\Yu-Gi-Oh!  Master Duel\LocalData\2509bcbc\0000' `
  -KeepDifferentVariants
```

Each bucket exports to a staging folder,
then reconciles into the output tree by file content:

- output missing: moved in as is.
- identical bytes: skipped.
- same path, different bytes: kept as `<name>_dYYYYMMDD<ext>`
  (for example `19375_d20260611.png`), so the new art sits next to the first.
  A same-day rerun that finds yet another distinct version adds a `_1`, `_2` suffix.

This relies on AssetStudio encoding the same texture to identical bytes each run,
which it does, so only genuinely different art produces a variant.
The mode re-extracts every file each run,
so it is slower than the default file-name skip.
Use it for cross-account comparison runs,
then review the `_dYYYYMMDD` files to see which cards differ between accounts.

After the export, move assets for account `dca6ade4` back to the original directory:

```bash
mv /d/game/SteamLibrary/steamapps/common/localdata_dca6ade4 "/d/game/SteamLibrary/steamapps/common/Yu-Gi-Oh!  Master Duel/LocalData/dca6ade4"
mv /d/game/SteamLibrary/steamapps/common/localsave_dca6ade4 "/d/game/SteamLibrary/steamapps/common/Yu-Gi-Oh!  Master Duel/LocalSave/dca6ade4"
```

#### Konami ID to card name

We will use script `cmd\rename_from_extracted\rename.go`
to rename PNG files from Konami card ID to card name.

Make sure the input and output directories in the code are correct before running it,
they look like this on Windows:

```
dirSourceBase = `D:\tmp_process_MD_file_by_path`
dirTargetBase = `D:\tmp_process_MD_file`
```

The other script `cmd\mix_dir_card_id\mix_dir_card_id.go`
is copy card art with name as Konami card ID to `D:\tmp_process_MD_file\card_id`,  
then they will be uploaded to [mdygo.daominah.uk](https://mdygo.daominah.uk/) (by WinSCP),  
to serve the Card Editor on [daominah.github.io](https://daominah.github.io/).

#### Compare and copy new arts to the final directory

After the renaming script, the output are in `D:\tmp_process_MD_file`,
we want to update the final output in `D:\syncthing\Master_Duel_art_full`.

We use [Meld](https://meldmerge.org/) to compare the two directories,
only copy the new updated arts to the final directory (avoid re-sync old arts).

```bash

# on Windows:

cd /d/syncthing/Master_Duel_art_full

# meld MD_art_renamed/ /d/tmp_process_MD_file/MD_art_renamed/
meld MD_different_censored/ /d/tmp_process_MD_file/MD_different_censored/
# meld MD_token_monster/ /d/tmp_process_MD_file/MD_token_monster/


# on Linux:

# cd /media/tungdt/WindowsData/syncthing/Master_Duel_art_full

# meld MD_art_renamed /media/tungdt/WindowsData/tmp_process_MD_file/MD_art_renamed
# meld MD_different_censored /media/tungdt/WindowsData/tmp_process_MD_file/MD_different_censored
# meld MD_token_monster /media/tungdt/WindowsData/tmp_process_MD_file/MD_token_monster

```

## Card text and effects

Card names and effect text are not in the images.
They are stored as `TextAsset` (`.bytes`) files under
`assets/resourcesassetbundle/card/data/<hash>/<locale>/`,
with one folder per locale (`en-us`, `ja-jp`, and a `md` metadata locale).

The `extract_md_assets_chunks.ps1` run already covers them through the `textAsset` type,
so excluding MonoBehaviour loses none of the card text.

The main files:

| File                    | Holds                                                      |
|-------------------------|------------------------------------------------------------|
| `CARD_Name`             | card names                                                 |
| `CARD_Desc`             | card effect / description text                             |
| `CARD_Indx`             | byte offsets that split Name/Desc blobs into per-card text |
| `CARD_Prop`             | properties: type, attribute, level, ATK/DEF                |
| `CARD_Genre`            | genre / archetype tags                                     |
| `CARD_Link`             | link arrows                                                |
| `CARD_RubyName`         | furigana reading of names                                  |
| `CARD_Same`             | alias / same-name card ids                                 |
| `WORD_Text`, `DLG_Text` | keyword glossary and tutorial dialog text                  |

The `.bytes` are obfuscated, not plain text.
Opening `CARD_Desc.bytes` directly shows garbage.
To read a card you de-obfuscate the `Name` and `Desc` blobs,
then use `CARD_Indx` offsets to slice out the string for each card id.

## Card image resolutions

The card illustration you collect has a single resolution per card:
`card/images/illust/{common,ocg,tcg}/<id>.png`, about 512 by 512 pixels.
There is no high or low variant inside the illust tree.

The game does keep lower-resolution copies of card faces named `<id>Low.png`,
but those live outside the illust tree,
embedded in the gameplay bundles that render a card small
(`cardfile`, `duel/effects`, `duel/timeline`, deck boxes).
They are a different crop and aspect (for example 128 by 256),
not a quality variant of the illustration,
so they are not part of the art set.

## Animations

There are two unrelated kinds of animation, stored and exported differently.

### Unity Animators (FBX)

Use `scripts/extract_md_animator.ps1`.
It runs the CLI in `-m animator` mode and exports FBX plus referenced textures.

- Almost all animators live in `masterduel_Data\data.unity3d` (about 2784),
  a single ~158 MB bundle that loads whole in about one minute.
- LocalData buckets hold only a few each, and StreamingAssets has none.
- This export cannot be chunked per hex bucket like the image script,
  because an animator references clips, mesh, and skeleton in other bundles.
  Each source is loaded whole instead.
- MonoBehaviour controller data needs IL2CPP dummy DLLs to be readable
  (pass `-AssemblyFolder`), but the FBX geometry and animation export without them.

The GUI's reputation for animator export errors comes from loading the whole
game folder at once. Exporting just `data.unity3d` runs clean.

### Monster cut-in (Spine 2D)

The summon cut-in animations are not Unity animators,
they are Spine 2D skeletal animations,
so they do not appear in the FBX export.
They are captured by the `textAsset` and `tex2d` types of the image extraction.

Location: `resourcesassetbundle/duel/timeline/duel/monstercutin/<region>/p<id>/`,
where `<region>` is `ocg` or `tcg`.

Each card ships two quality tiers, `highend_hd` and `sd`.
Inside each tier the atlas and texture pages sit in a scale-named subfolder,
while the skeleton and mesh sit at the tier root:

```text
p<id>/
  highend_hd/
    P<id>JS.asset                          # skeleton, or P<id>JS.json (see note)
    Skeleton Prefab Mesh [SpineP<id>].obj
    <scale>/                               # e.g. 0.56 or 0.97
      P<id>.atlas
      P<id>.png, P<id>_2.png, ...
  sd/
    ... same layout, smaller scale ...
```

The scale and the number and size of texture pages vary per card,
they are not fixed per tier. Two examples:

| Card  | Tier         | Scale | Texture pages     |
|-------|--------------|-------|-------------------|
| 19375 | `highend_hd` | 0.56  | four 2048 by 2048 |
| 19375 | `sd`         | 0.28  | two 2048 by 2048  |
| 10001 | `highend_hd` | 0.97  | one 4096 by 4096  |
| 10001 | `sd`         | 0.485 | one 2048 by 2048  |

The files in each tier:

- `P<id>JS.json` or `P<id>JS.asset`: the Spine skeleton and animation,
  Spine version 4.2.20.
  The `.json` versus `.asset` extension does not track the tier reliably:
  some cards use `.asset` in both tiers, some use `.json` in both,
  some split by tier. It depends on the card and region.
- `P<id>.atlas`: maps skeleton attachments to regions of the texture pages.
- `P<id>.png`, `P<id>_2.png`, ...: the atlas pages.
- `Skeleton Prefab Mesh [SpineP<id>].obj`: a static bind-pose mesh only,
  no animation, useful only to see the shape.
  The bracketed name spacing varies, for example `[SpineP<id>]` or `[Spine P<id>]`.

The skeleton stores a native canvas size, for example Centur-Ion Legatia (19375)
is 4155 by 4178 at 60 fps.
The textures ship downscaled (card 19375 `highend_hd` tier at 0.56),
so the sharpest real detail is around 2327 by 2340,
and rendering at full native upscales the textures.

To view and screenshot a cut-in,
load `P<id>JS.json` plus `P<id>.atlas` plus the `.png` pages
in the Official Spine editor (trial),
which plays Spine 4.2 natively and has a frame-accurate timeline.
Blender has no native Spine support and is not a usable route.

## Remaining work

- View the monster cut-in animations in the Official Spine editor (trial),
  using the `highend_hd` tier files for one card as a first test.
- Add a script that collects a single card's cut-in assets by id
  (skeleton json, atlas, texture pages) into one folder,
  named and laid out so the Spine editor can open it directly.
- Verify OCG coverage and collect the missing OCG art.
  OCG counts taken from the merged `D:\tmp_process_MD_file_by_path\assets`
  working tree are unreliable for this:
  much of that OCG art originated from the Japanese account on another machine,
  not from a clean `70102374` extraction.
  To measure real coverage, extract each account in isolation to a fresh folder,
  `LocalData\70102374` (English) and `LocalData\2509bcbc`
  (Japanese, present but not yet parsed on this machine),
  then compare their OCG id sets.
  Alternatively, run the Japanese account with `-KeepDifferentVariants`
  (see the chunked export section) against the existing tree:
  any card whose Japanese art differs lands as a `_dYYYYMMDD` sibling,
  which directly surfaces the cross-account differences.

## Result

All arts result are published at:

* https://mdygo.daominah.uk/  (powered by Cloudflared Tunnel)
* https://mdygo2048.daominah.uk/
* [Google Drive daominah@gmail.com](https://drive.google.com/drive/folders/1PVaWUaullSjaSKwbOi3Q1Xj024Qzq4YD?usp=share_link) (not updated frequently)

Example a card, `Ryzeal Detonator`,  
game file `20578.png`,  
renamed to `ryzeal_detonator_20578.png`.

## References

* [Guide from Reddit](https://www.reddit.com/r/masterduel/comments/uszzul/guide_to_create_card_art_replacement_file_pc/)
* [Konami official cards database](https://www.db.yugioh-card.com/yugiohdb/card_search.action?ope=2&cid=4007&request_locale=en)
* [ygocdb.com for search YuGiOh card ID](https://ygocdb.com/)
* [Install Golang](https://golang.org/doc/install)
* [OCG art uncensored](https://www.youtube.com/watch?v=hXGVXXHT6us)
