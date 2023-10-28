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
	"strings"

	yugioh "github.com/daominah/yugioh_master_duel_card_art"
)

var (
	dirSourceCardArt string
	dirTargetCardArt string
	targetNameSuffix string
)

var weirdDir = `/media/tungdt/WindowsData/syncthing/Master_Duel_art_full/Texture2D_weird`

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

	// MD_file_Japanese uncensored OCG arts will be handled in `cmd/diff_texture/diff_texture.go`,
	// this `rename.go` only renames English art.
	flag.StringVar(&dirSourceCardArt, "dirSourceCardArt",
		`/media/tungdt/WindowsData/syncthing/Master_Duel_art_full/MD_file_English/Texture2D`,
		"path to source directory that contains extracted images from the game")
	flag.StringVar(&dirTargetCardArt, "dirTargetCardArt",
		`/media/tungdt/WindowsData/syncthing/Master_Duel_art_full/art_renamed_all`,
		"path to target directory that contains output renamed images from this program")
	flag.StringVar(&targetNameSuffix, "targetNameSuffix",
		``,
		`set to "_ocg" if processing Japanese arts to append to output files name`)

	flag.Parse()
	log.Printf("flag vars:")
	log.Printf("dirSourceCardArt: %v", dirSourceCardArt)
	log.Printf("dirTargetCardArt: %v", dirTargetCardArt)

	cards := yugioh.ReadAllCardData()

	log.Printf("doing read dir %v", dirSourceCardArt)
	dir, err := os.ReadDir(dirSourceCardArt)
	if err != nil {
		log.Fatalf("error os.ReadDir: %v", err)
	}
	nCopiedFiles := 0
	for i, f := range dir {
		//if i > 100 { // small number for testing
		//	break
		//}
		sourceFullPath := filepath.Join(dirSourceCardArt, f.Name())
		nameNoExt := strings.TrimSuffix(f.Name(), ".png")
		cardInfo, found := cards[nameNoExt]
		if !found {
			if false {
				sourceFile, err := os.Open(sourceFullPath)
				if err != nil {
					log.Printf("error os.ReadFile: %v", err)
					continue
				}
				targetFullPath := filepath.Join(weirdDir, f.Name())
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
				log.Printf("created missing info card %v nCopiedBytes %v", f.Name(), nCopiedBytes)
			} else {
				log.Printf("i %v ignore %v", i, f.Name())
			}
			continue
		}

		enName := cardInfo.EnName
		if enName == "" {
			enName = cardInfo.WikiEn
		}
		targetName := fmt.Sprintf("%v_%v%v.png",
			yugioh.NormalizeName(enName), cardInfo.Cid, targetNameSuffix)

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
