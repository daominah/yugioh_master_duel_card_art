package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	yugioh "github.com/daominah/yugioh_master_duel_card_art"
)

// this program read all images with name "{cardID}.png"
// (extracted from YuGiOh Master Duel, see README.md for more info),
// then copy them to a new directory with name "{cardEnglishName}.png";
func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

	// the following dir and file paths only work on Linux
	const (
		dirBase = "/media/tungdt/WindowsData/syncthing/Master_Duel_art_full"

		dirSourceCardArtCommon = dirBase + "/MD_file/assets/resources/card/images/illust/common"
		dirSourceCardArtOCG    = dirBase + "/MD_file/assets/resources/card/images/illust/ocg"
		dirSourceCardArtTCG    = dirBase + "/MD_file/assets/resources/card/images/illust/tcg"

		dirTargetCardArt = dirBase + "/MD_art_renamed"
		dirTokenMonster  = dirBase + "/MD_token_monster"
		dirDiffCensor    = dirBase + "/MD_different_censored"
	)

	log.Printf("_______________________________________________________")
	for _, v := range []string{dirTargetCardArt, dirTokenMonster, dirDiffCensor} {
		if _, err := os.Stat(v); err != nil {
			log.Fatalf("error probably directory does not exist: %v", v)
		}
	}
	log.Printf("all directory exists, READY to process")
	time.Sleep(1 * time.Second)

	cards := yugioh.ReadAllCardDataKonami()

	nProcessed := 0
	nCopiedFiles := 0
	nCopiedFilesCensor := 0
	for _, dirSourceCardArt := range []string{dirSourceCardArtCommon, dirSourceCardArtOCG, dirSourceCardArtTCG} {
		dir, err := os.ReadDir(dirSourceCardArt)
		if err != nil {
			log.Fatalf("error os.ReadDir: %v", err)
		}
		for _, subDir := range dir {
			subDirPath := filepath.Join(dirSourceCardArt, subDir.Name())
			subDir, err := os.ReadDir(subDirPath)
			if err != nil {
				log.Printf("error os.ReadDir: %v", err)
			}
			for _, f := range subDir {
				nProcessed += 1
				log.Printf("i: %-5v, f: %v", nProcessed, f.Name())
				sourceFullPath := filepath.Join(subDirPath, f.Name())
				cardID, fragment := getCardIDFromFileName(f.Name())
				if fragment != "" {
					// after changed Export Option in Asset Studio to group by file path
					// fragment images are usually duplicated
					continue
				}
				cardInfo, found := cards[cardID]

				if !found {
					// OCG exclusive cards or new cards that have not been in TCG, or Token;
					// this code section moves them to "dirTokenMonster"
					maybeCardID, err := strconv.Atoi(cardID)
					if err != nil {
						continue
					}
					if maybeCardID < 3000 || maybeCardID > 50000 {
						// Konami cardID start from 4007 Blue-Eyes White Dragon,
						// some alternative arts have cardID between 3000-4000
						continue
					}
					sourceFile, err := os.Open(sourceFullPath)
					if err != nil {
						log.Printf("error os.ReadFile: %v", err)
						continue
					}
					targetFullPath := filepath.Join(dirTokenMonster, f.Name())
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

				targetName := fmt.Sprintf("%v_%v", yugioh.NormalizeName(cardInfo.CardName), cardInfo.CardID)
				if cardInfo.AltArtID != "" {
					targetName += "_alt" + cardInfo.AltArtID
				}

				// sometimes Asset Studio save OCG art 1st (without fragment), sometimes TCG art 1st
				needCopyToDirDiffCensor := false
				switch dirSourceCardArt {
				case dirSourceCardArtOCG:
					targetName += "_ocg"
					needCopyToDirDiffCensor = true
				case dirSourceCardArtTCG:
					targetName += "_tcg"
					needCopyToDirDiffCensor = true
				default:
					// do nothing
				}
				targetName += ".png"

				targetFullPath := filepath.Join(dirTargetCardArt, targetName)
				isCopied := copyFile(sourceFullPath, targetFullPath)
				if isCopied {
					nCopiedFiles += 1
					log.Printf("created new file %v", targetName)
				}

				if needCopyToDirDiffCensor {
					targetFullPath := filepath.Join(dirDiffCensor, targetName)
					isCopied := copyFile(sourceFullPath, targetFullPath)
					if isCopied {
						nCopiedFilesCensor += 1
						log.Printf("created new file in dirDiffCensor %v", targetName)
					}
				}
			}
		}
	}

	log.Printf("-------------------------------------------------------")
	log.Printf("-------------------------------------------------------")
	log.Printf("func main returned")
	log.Printf("nCopiedFiles: %v", nCopiedFiles)
	log.Printf("nCopiedFilesCensor: %v", nCopiedFilesCensor)
}

// getCardIDFromFileName returns (cardID, fragment),
// examples:
// "14937.png" (Selene Queen of the Master Magicians) returns ("14937", "")
// "14937 #400847.png" (Selene's art too but probably for TCG) returns ("14937", "400847")
func getCardIDFromFileName(fileName string) (string, string) {
	nameNoExt := strings.TrimSuffix(fileName, ".png")
	cardID, fragment := nameNoExt, "" // fragment is number after # sign in file name
	if tmp := strings.Index(nameNoExt, "#"); tmp != -1 {
		cardID = strings.TrimSpace(nameNoExt[:tmp])
		fragment = strings.TrimSpace(nameNoExt[tmp+1:])
	}
	return cardID, fragment
}

// copyFile logs and handles error too
func copyFile(sourceFullPath string, targetFullPath string) bool {
	sourceFile, err := os.Open(sourceFullPath)
	if err != nil {
		log.Printf("error os.ReadFile: %v", err)
		return false
	}
	if _, err := os.Stat(targetFullPath); err == nil {
		//log.Printf("do nothing because of existed %v", targetFullPath)
		return false
	}
	targetFile, err := os.Create(targetFullPath)
	if err != nil {
		log.Printf("error os.Create: %v", err)
		return false
	}
	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		log.Printf("error io.Copy: %v", err)
		return false
	}
	return true
}
