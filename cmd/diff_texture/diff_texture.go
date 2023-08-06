package main

import (
	"flag"
	"fmt"
	yugioh "github.com/daominah/yugioh_master_duel_card_art"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	dirEnglish     string // Master Dual English arts
	dirJapanese    string // Master Dual Japanese arts
	dirDifferent   string // pairs of English and Japanese art if they are different
	dirDiffRenamed string // Japanese art renamed card id to card name
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)
	flag.StringVar(&dirEnglish, "englishFileDir",
		`/media/tungdt/WindowsData/syncthing/Master_Duel_art_full/MD_file_English/Texture2D`, "")
	flag.StringVar(&dirJapanese, "dirJapanese",
		`/media/tungdt/WindowsData/syncthing/Master_Duel_art_full/MD_file_Japanese/Texture2D`, "")
	flag.StringVar(&dirDifferent, "dirDifferent",
		`/media/tungdt/WindowsData/syncthing/Master_Duel_art_full/MD_file_different`, "")
	flag.StringVar(&dirDiffRenamed, "dirDiffRenamed",
		`/media/tungdt/WindowsData/syncthing/Master_Duel_art_full/art_renamed_different`, "")
	log.Printf("dirEnglish: %v\n", dirEnglish)
	log.Printf("dirJapanese: %v\n", dirJapanese)

	dirEnglishObj, err := os.ReadDir(dirEnglish)
	if err != nil {
		log.Fatalf("error os.ReadDir: %v", err)
	}
	cards := yugioh.ReadAllCardData()
	for i, f := range dirEnglishObj {
		log.Printf("i %v ____________________________________________\n", i)
		//if i > 100 {
		//	break
		//}

		ext := filepath.Ext(f.Name()) // ext includes the dot "."
		cid := strings.TrimSuffix(f.Name(), ext)
		engFullPath := filepath.Join(dirEnglish, f.Name())
		jpnFullPath := filepath.Join(dirJapanese, f.Name())
		diffFullPathEng := filepath.Join(dirDifferent, f.Name())
		diffFullPathJpn := filepath.Join(dirDifferent, cid+"_ocg"+ext)
		cardInfo, found := cards[cid]
		var cardName string
		if found {
			enName := cardInfo.EnName
			if enName == "" {
				enName = cardInfo.WikiEn
			}
			cardName = fmt.Sprintf("%v_%v_%v",
				yugioh.NormalizeName(enName), cardInfo.Id, cardInfo.Cid)
		}
		engFile, err := f.Info()
		if err != nil {
			log.Printf("error engFile info: %v", err)
			continue
		}
		jpnFile, err := os.Stat(jpnFullPath)
		if err != nil {
			log.Printf("jpnFile Stat: %v", err)
			continue
		}
		if engFile.Size() == jpnFile.Size() { // TODO: should check by sha256sum
			log.Printf("same file size, skip %v", f.Name())
			continue
		}
		for k, srcFullPath := range []string{engFullPath, jpnFullPath} {
			var target1, target2 string
			if k == 0 {
				target1 = diffFullPathEng
				if cardName != "" {
					target2 = filepath.Join(dirDiffRenamed, cardName+ext)
				}
			} else if k == 1 {
				target1 = diffFullPathJpn
				if cardName != "" {
					target2 = filepath.Join(dirDiffRenamed, cardName+"_ocg"+ext)
				}
			}
			sourceFile, err := os.Open(srcFullPath)
			if err != nil {
				log.Printf("error os.ReadFile: %v", err)
				continue
			}
			targets := []string{target1}
			if target2 != "" {
				targets = []string{target1, target2}
			}
			for _, target := range targets {
				if _, err := os.Stat(target); err == nil {
					log.Printf("do nothing because of existed %v", target)
					continue
				}
				targetFile, err := os.Create(target)
				if err != nil {
					log.Printf("error os.Create: %v", err)
					continue
				}
				_, err = io.Copy(targetFile, sourceFile)
				if err != nil {
					log.Printf("error io.Copy: %v", err)
					continue
				}
				targetFile.Close()
				log.Printf("ok copied to %v", target)
			}
		}
	}
}
