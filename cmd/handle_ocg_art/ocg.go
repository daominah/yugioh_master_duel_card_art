package main

import (
	"bytes"
	"flag"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"os"

	yugioh "github.com/daominah/yugioh_master_duel_card_art"
	"golang.org/x/image/bmp"
)

var (
	dirSourceCardArt string
	dirTargetCardArt string
)

func init() {
	flag.StringVar(&dirSourceCardArt, "dirSourceCardArt",
		`/home/tungdt/opt/card_Master_Duel/ocg_art_uncensored_bmp`,
		"path to source directory")
	flag.StringVar(&dirTargetCardArt, "dirTargetCardArt",
		`/home/tungdt/opt/card_Master_Duel/ocg_art_uncensored`,
		"path to target directory")
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

	// TODO: run in goroutines because func ConvertBmpToPng is slow
	nCopiedFiles := 0
	for i, f := range dir {
		//if i >= 4 { // small number for testing
		//	break
		//}
		sourceFullPath := filepath.Join(dirSourceCardArt, f.Name())
		words := strings.Split(f.Name(), " ")
		if len(words) < 2 { // card name format: "{cardId} {cen|unc} - name.bmp"
			log.Printf("ignore unexpected file name %v", f.Name())
			continue
		}
		cardId := words[0]
		var ocgSuffix string
		if words[1] != "(cen)" {
			ocgSuffix = "ocg"
		}

		cardInfo, found := cards[cardId]
		if !found {
			log.Printf("i %v ignore %v", i, f.Name())
			continue
		}
		enName := cardInfo.EnName
		if enName == "" {
			enName = cardInfo.WikiEn
		}
		targetName := fmt.Sprintf("%v_%v_%v_%v.png",
			yugioh.NormalizeName(enName), cardInfo.Id, cardInfo.Cid, ocgSuffix)

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

		bs, err := io.ReadAll(sourceFile)
		if err != nil {
			log.Printf("error io.ReadAll: %v", err)
			continue
		}
		pngBytes, err := ConvertBmpToPng(bs)
		if err != nil {
			log.Printf("error ConvertBmpToPng: %v", err)
			continue
		}
		_, err = io.Copy(targetFile, bytes.NewReader(pngBytes))
		if err != nil {
			log.Printf("error io.Copy: %v", err)
			continue
		}
		nCopiedFiles += 1
		log.Printf("created new file %v", targetName)
	}
	log.Printf("func main returned, nCopiedFiles: %v", nCopiedFiles)
}

func ConvertBmpToPng(imageBytes []byte) ([]byte, error) {
	imgType := http.DetectContentType(imageBytes)
	switch imgType {
	case "image/png":
		return imageBytes, nil
	case "image/jpeg":
		img, err := jpeg.Decode(bytes.NewReader(imageBytes))
		if err != nil {
			return nil, fmt.Errorf("decode jpeg: %v", err)
		}
		buf := new(bytes.Buffer)
		if err := png.Encode(buf, img); err != nil {
			return nil, fmt.Errorf("encode png: %v", err)
		}
		return buf.Bytes(), nil
	case "image/bmp":
		img, err := bmp.Decode(bytes.NewReader(imageBytes))
		if err != nil {
			return nil, fmt.Errorf("decode jpeg: %v", err)
		}
		buf := new(bytes.Buffer)
		if err := png.Encode(buf, img); err != nil {
			return nil, fmt.Errorf("encode png: %v", err)
		}
		return buf.Bytes(), nil
	default:
		return nil, fmt.Errorf("unsupported image type %v", imgType)
	}
}
