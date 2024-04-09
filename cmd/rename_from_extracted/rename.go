// this program read all images with name "{cardID}.png" (extracted from YuGiOh Master Duel)
// then copy them to a new directory with name "{cardEnglishName}.png"
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	yugioh "github.com/daominah/yugioh_master_duel_card_art"
)

var (
	dirSourceCardArt string
	dirTargetCardArt string
)

// dirCardsNoData contains images that have name (cardID) cannot be found on Konami db,
// example "15067.png" (is just duplicated art of "15036.png")
var dirCardsNoData = `/media/tungdt/WindowsData/syncthing/Master_Duel_art_full/MD_card_no_data`

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

	flag.StringVar(&dirSourceCardArt, "dirSourceCardArt",
		`/media/tungdt/WindowsData/syncthing/Master_Duel_art_full/MD_file_Japanese/Texture2D`,
		"path to source directory that contains extracted images from the game")
	flag.StringVar(&dirTargetCardArt, "dirTargetCardArt",
		`/media/tungdt/WindowsData/syncthing/Master_Duel_art_full/MD_art_renamed`,
		"path to target directory that contains output renamed images from this program")
	flag.Parse()
	log.Printf("flag vars:")
	log.Printf("dirSourceCardArt: %v", dirSourceCardArt)
	log.Printf("dirTargetCardArt: %v", dirTargetCardArt)

	cards := yugioh.ReadAllCardDataKonami()

	log.Printf("doing read dir %v", dirSourceCardArt)
	dir, err := os.ReadDir(dirSourceCardArt)
	if err != nil {
		log.Fatalf("error os.ReadDir: %v", err)
	}
	nCopiedFiles := 0
	for i, f := range dir {
		sourceFullPath := filepath.Join(dirSourceCardArt, f.Name())
		nameNoExt := strings.TrimSuffix(f.Name(), ".png")
		cardInfo, found := cards[nameNoExt]
		if !found {
			maybeCardID, err := strconv.Atoi(nameNoExt)
			if err != nil {
				continue
			}
			if maybeCardID < 3000 || maybeCardID > 30000 {
				// Konami cardID start from 4007 Blue-Eyes White Dragon,
				// some alternative arts have cardID about 3000-4000
				continue
			}
			sourceFile, err := os.Open(sourceFullPath)
			if err != nil {
				log.Printf("error os.ReadFile: %v", err)
				continue
			}
			targetFullPath := filepath.Join(dirCardsNoData, f.Name())
			if _, err := os.Stat(targetFullPath); err == nil {
				continue
			}
			targetFile, err := os.Create(targetFullPath)
			if err != nil {
				log.Printf("error os.Create: %v", err)
				continue
			}
			nCopiedBytes, err := io.Copy(targetFile, sourceFile)
			if err != nil {
				log.Printf("error io.Copy: %v", err)
				continue
			}
			log.Printf("saved missing info card %v nCopiedBytes %v", f.Name(), nCopiedBytes)

			continue
		}

		targetName := fmt.Sprintf("%v_%v.png",
			yugioh.NormalizeName(cardInfo.CardName), cardInfo.CardID)
		if cardInfo.AltArtID != "" {
			targetName = fmt.Sprintf("%v_%v_alt%v.png",
				yugioh.NormalizeName(cardInfo.CardName), cardInfo.CardID, cardInfo.AltArtID)
		}
		targetFullPath := filepath.Join(dirTargetCardArt, targetName)
		log.Printf("i %v doing copy `%v` to `%v`", i, f.Name(), targetName)

		sourceFile, err := os.Open(sourceFullPath)
		if err != nil {
			log.Printf("error os.ReadFile: %v", err)
			continue
		}
		if _, err := os.Stat(targetFullPath); err == nil {
			log.Printf("do nothing because of existed %v", targetFullPath)
			continue
		}
		targetFile, err := os.Create(targetFullPath)
		if err != nil {
			log.Printf("error os.Create: %v", err)
			continue
		}
		nCopiedBytes, err := io.Copy(targetFile, sourceFile)
		if err != nil {
			log.Printf("error io.Copy: %v", err)
			continue
		}
		nCopiedFiles += 1
		log.Printf("created new file %v nCopiedBytes %v", targetName, nCopiedBytes)
	}
	log.Printf("func main returned, nCopiedFiles: %v", nCopiedFiles)
}
