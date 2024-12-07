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

// this program read all images with name "{cardID}.png" or "{cardID} #{fragment}.png"
// (extracted from YuGiOh Master Duel, see README.md for more info),
// then copy them to a new directory "${dirTargetCardArt}",
// with name changed to "{cardEnglishName}.png";
func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

	// constants,
	// these are paths to input and output directories,
	// depending on Linux or Windows, change them before running this script,
	// do not change them while executing this program
	var (
		// dirSourceBase = "/media/tungdt/WindowsData/tmp_process_MD_file_by_path"
		dirSourceBase = `D:\tmp_process_MD_file_by_path`

		// dirTargetBase = "/media/tungdt/WindowsData/tmp_process_MD_file"
		dirTargetBase = `D:\tmp_process_MD_file`

		// TODO: handle sources: resourcesassetbundle, resourcesassetbundleintutorial

		dirSourceCardArtCommon = filepath.Join(dirSourceBase, "/assets/resources/card/images/illust/common")
		dirSourceCardArtOCG    = filepath.Join(dirSourceBase, "/assets/resources/card/images/illust/ocg")
		dirSourceCardArtTCG    = filepath.Join(dirSourceBase, "/assets/resources/card/images/illust/tcg")

		dirTargetCardArt = filepath.Join(dirTargetBase, "/MD_art_renamed")
		dirTokenMonster  = filepath.Join(dirTargetBase, "/MD_token_monster")
		dirDiffCensor    = filepath.Join(dirTargetBase, "/MD_different_censored")
	)

	log.Printf("_______________________________________________________")
	for _, v := range []string{dirTargetCardArt, dirTokenMonster, dirDiffCensor} {
		if _, err := os.Stat(v); err != nil {
			log.Fatalf("error probably directory does not exist: %v", v)
			// mkdir MD_art_renamed MD_different_censored MD_token_monster
		}
	}
	log.Printf("all target directories exist, READY to process")
	time.Sleep(1 * time.Second)

	cards := yugioh.ReadAllCardDataKonami()

	beginProcessT := time.Now()
	nProcessed := 0
	nCopiedAsTokenCards := 0
	nCopiedFiles := 0
	nCopiedFilesCensor := 0

	for _, dirSourceCardArt := range []string{dirSourceCardArtCommon, dirSourceCardArtOCG, dirSourceCardArtTCG} {
		cardIDPrefixDirs, err := os.ReadDir(dirSourceCardArt) // return ["03", "04", ..., "20"]
		if err != nil {
			log.Fatalf("error os.ReadDir: %v", err)
		}
		log.Printf("doing directory %v", dirSourceCardArt)
		time.Sleep(1 * time.Second)
		for _, subDir := range cardIDPrefixDirs {
			subDirPath := filepath.Join(dirSourceCardArt, subDir.Name())
			subDir, err := os.ReadDir(subDirPath)
			if err != nil {
				log.Printf("error os.ReadDir: %v", err)
			}
			for _, f := range subDir { // f is in ["4007.png", "4008.png", ..., "4999.png"], Blue-Eyes to Osiris
				nProcessed += 1
				log.Printf("nProcessed: %-5v, f: %v", nProcessed, f.Name())
				sourceFullPath := filepath.Join(subDirPath, f.Name())
				cardID, fragment := getCardIDFromFileName(f.Name())
				if fragment != "" {
					// after changed Export Option in Asset Studio to group by file path
					// fragment images are usually duplicated
					continue
				}
				cardInfo, found := cards[cardID]

				if !found { // minority case, mostly found is true
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
						// target file existed
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
					nCopiedAsTokenCards += 1
					log.Printf("saved missing info card %v nCopiedBytes %v", f.Name(), nCopiedBytes)
					continue
				}

				targetName := fmt.Sprintf("%v_%v", yugioh.NormalizeName(cardInfo.CardName), cardInfo.CardID)
				if cardInfo.AltArtID != "" {
					targetName += "_alt" + cardInfo.AltArtID
				}

				// sometimes Asset Studio save OCG art 1st (without fragment), sometimes TCG art 1st
				needCopyToDirCensor := false
				switch dirSourceCardArt {
				case dirSourceCardArtOCG:
					targetName += "_ocg"
					needCopyToDirCensor = true
				case dirSourceCardArtTCG:
					targetName += "_tcg"
					needCopyToDirCensor = true
				default:
					// do nothing
				}
				targetName += ".png"
				targetFullPath := filepath.Join(dirTargetCardArt, targetName)
				// copyFile does nothing if target file existed
				isCopied := copyFile(sourceFullPath, targetFullPath)
				if isCopied {
					nCopiedFiles += 1
					log.Printf("created new file %v", targetName)
				}

				if needCopyToDirCensor {
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
	log.Printf("done processing %v files, duration: %v", nProcessed, time.Since(beginProcessT))
	log.Printf("nCopiedFiles: %v", nCopiedFiles)
	log.Printf("nCopiedFilesCensor: %v", nCopiedFilesCensor)
	log.Printf("nCopiedAsTokenCards: %v", nCopiedAsTokenCards)
	log.Printf("check the target directories for results:\n%v\n%v\n%v", dirTargetCardArt, dirTokenMonster, dirDiffCensor)
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

// copyFile does nothing if target file existed,
// copyFile logs and handles error too
func copyFile(sourceFullPath string, targetFullPath string) bool {
	sourceFile, err := os.Open(sourceFullPath)
	if err != nil {
		log.Printf("error os.ReadFile: %v", err)
		return false
	}
	if _, err := os.Stat(targetFullPath); err == nil {
		// log.Printf("do nothing because target file existed %v", targetFullPath)
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
