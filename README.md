# YuGiOh Master Duel all cards art

Extract cards image data from "Yu-Gi-Oh! Master Duel" and rename files from card id to card name.

### Steps

* Download the game from Steam on Windows:  
  [steampowered.com/app/../YuGiOh_Master_Duel](https://store.steampowered.com/app/1449850/YuGiOh_Master_Duel/)

* Download app `AssetStudio` (to extract image data from the game)  
  [github.com/Perfare/AssetStudio](https://github.com/Perfare/AssetStudio/releases)  
  Default .NET verison on Windows is v4.x

* Use `AssetStudio` to open directory
  `steamapps/common/Yu-Gi-Oh!  Master Duel/LocalData` in the Steam library.
  It takes a long duration and use a lot of RAM (16GB is enough).
  Save all files with type "Texture2D", output is a lot of PNG images (size about 12GB)
  
* Before run `cmd/rename_from_extracted/rename.go` to rename PNG files from card ID to card name,
  need to change flag `dirSourceCardArt` and `dirTargetCardArt` to right directories.

### Result

Example card `Blue-Eyes Alternative White Dragon`, game file `12253.png`,
  renamed to `blue_eyes_alternative_white_dragon_38517737_12253.png`:
  ![Blue-Eyes Alternative White Dragon](12253.png)

All arts result, uploaded on [Google drive](https://drive.google.com/drive/folders/1PVaWUaullSjaSKwbOi3Q1Xj024Qzq4YD?usp=share_link) (public read) 

### References

* [Guide from Reddit](https://www.reddit.com/r/masterduel/comments/uszzul/guide_to_create_card_art_replacement_file_pc/)
* [Website for search YuGiOh card ID](https://ygocdb.com/)
* [Install Golang](https://golang.org/doc/install)
* [OCG art uncensored](https://www.youtube.com/watch?v=hXGVXXHT6us)
