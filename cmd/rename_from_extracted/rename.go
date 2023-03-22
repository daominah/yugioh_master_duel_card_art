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
)

func init() {
	flag.StringVar(&dirSourceCardArt, "dirSourceCardArt",
		`/media/tungdt/WindowsData/picture/card_Master_Duel_art`,
		"path to source directory that contains extracted images from the game")
	flag.StringVar(&dirTargetCardArt, "dirTargetCardArt",
		`/home/tungdt/opt/card_Master_Duel/in_game_art_renamed`,
		"path to target directory that contains output renamed images from this program")
}

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

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
		fnameWOExt := strings.TrimSuffix(f.Name(), ".png")
		cardInfo, found := cards[fnameWOExt]
		if !found {
			log.Printf("i %v ignore %v", i, f.Name())
			continue
		}
		enName := cardInfo.EnName
		if enName == "" {
			enName = cardInfo.WikiEn
		}
		targetName := fmt.Sprintf("%v_%v_%v.png", yugioh.NormalizeName(enName), cardInfo.Id, cardInfo.Cid)
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
