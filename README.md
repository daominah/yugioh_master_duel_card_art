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
* Filter Type: choose all except the following types:
  - MonoBehaviour (a lot of human unreadable files)
  - Animator (probably they cause exporting errors)
* Export: Filtered assets: choose output to `D:\tmp_process_MD_file_by_path`
  (make sure the directory exists and empty).

This takes about 3 hours,
occasionally show errors and stuck, require human to click OK.

The result card arts as PNG, named as Konami card ID,
most are in dir `assets/resources/card/images/illust/common`.

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
