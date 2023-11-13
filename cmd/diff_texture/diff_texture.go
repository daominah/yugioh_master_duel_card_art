package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	yugioh "github.com/daominah/yugioh_master_duel_card_art"
)

var (
	dirEnglish     string // Master Dual English arts
	dirJapanese    string // Master Dual Japanese arts
	dirDifferent   string // pairs of MD files (English, Japanese) art if they are different
	dirDiffRenamed string
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)
	flag.StringVar(&dirEnglish, "englishFileDir",
		`/media/tungdt/WindowsData/syncthing/Master_Duel_art_full/MD_file_English/Texture2D`, "")
	flag.StringVar(&dirJapanese, "dirJapanese",
		`/media/tungdt/WindowsData/syncthing/Master_Duel_art_full/MD_file_Japanese/Texture2D`, "")
	flag.StringVar(&dirDifferent, "dirDifferent",
		`/media/tungdt/WindowsData/syncthing/Master_Duel_art_full/different_MD_file2`, "")
	flag.StringVar(&dirDiffRenamed, "dirDiffRenamed",
		`/media/tungdt/WindowsData/syncthing/Master_Duel_art_full/different_renamed2`, "")
	log.Printf("dirEnglish: %v\n", dirEnglish)
	log.Printf("dirJapanese: %v\n", dirJapanese)
	log.Printf("dirDifferent: %v\n", dirDifferent)
	log.Printf("dirDiffRenamed: %v\n", dirDiffRenamed)

	dirEnglishObj, err := os.ReadDir(dirEnglish)
	if err != nil {
		log.Fatalf("error os.ReadDir: %v", err)
	}
	for _, v := range []string{dirDifferent, dirDiffRenamed} {
		_, err := os.ReadDir(v)
		if err != nil {
			log.Fatalf("error os.ReadDir: %v", err)
		}
	}
	cards := yugioh.ReadAllCardData()

	beginT := time.Now()
	limitGoroutines := make(chan bool, 8)
	wg := &sync.WaitGroup{}
	for i, f := range dirEnglishObj {
		//if i > 100 {
		//	break
		//}
		limitGoroutines <- true
		wg.Add(1)
		go func(i int, f os.DirEntry) {
			defer func() {
				<-limitGoroutines
				wg.Add(-1)
			}()
			log.Printf("i %v ____________________________________________\n", i)
			ext := filepath.Ext(f.Name())            // ext includes the dot "."
			cid := strings.TrimSuffix(f.Name(), ext) // file name is Konami cardID
			engFullPath := filepath.Join(dirEnglish, f.Name())
			japFullPath := filepath.Join(dirJapanese, f.Name())
			diffFullPathEng := filepath.Join(dirDifferent, f.Name())
			diffFullPathJap := filepath.Join(dirDifferent, cid+"_ocg"+ext)
			cardInfo, found := cards[cid]
			var cardName string
			if found {
				enName := cardInfo.EnName
				if enName == "" {
					enName = cardInfo.WikiEn
				}
				cardName = fmt.Sprintf("%v_%v",
					yugioh.NormalizeName(enName), cardInfo.Cid)
			}

			engBytes, err := os.ReadFile(engFullPath)
			if err != nil {
				log.Printf("error os.ReadFile %v: %v", engFullPath, err)
				return
			}
			japBytes, err := os.ReadFile(japFullPath)
			if err != nil {
				//log.Printf("skip because file name not found in Japanese dir: %v", f.Name())
				return
			}
			if bytes.Compare(engBytes, japBytes) == 0 {
				//log.Printf("skip because of same file content: %v", f.Name())
				return
			}

			for k, srcBytes := range [][]byte{engBytes, japBytes} {
				var target1, target2 string
				if k == 0 {
					target1 = diffFullPathEng
					if cardName != "" {
						target2 = filepath.Join(dirDiffRenamed, cardName+ext)
					}
				} else if k == 1 {
					target1 = diffFullPathJap
					if cardName != "" {
						target2 = filepath.Join(dirDiffRenamed, cardName+"_ocg"+ext)
					}
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
					_, err = io.Copy(targetFile, bytes.NewReader(srcBytes))
					if err != nil {
						log.Printf("error io.Copy: %v", err)
						continue
					}
					_ = targetFile.Close()
					log.Printf("copied to %v", target)
				}
			}
		}(i, f)
	}
	wg.Wait()
	log.Printf("main diff_texture returned, loop duration: %v", time.Since(beginT))
}
