package main

import (
	"log"
	"os"

	yugioh "github.com/daominah/yugioh_master_duel_card_art"
)

func main() {
	validNames := yugioh.NewTrie()
	for _, card := range yugioh.ReadAllCardDataKonami() {
		validNames.Insert(yugioh.NormalizeName(card.CardName))
	}

	targetDir := `D:\syncthing\Master_Duel_art_full\upscayled_2048`
	// read targetDir then print file name
	dirEntries, err := os.ReadDir(targetDir)
	if err != nil {
		log.Fatalf("error os.ReadDir: %v", err)
	}
	for _, fileObj := range dirEntries {
		if fileObj.IsDir() {
			continue
		}
		if !validNames.CheckPrefixIsAKey(fileObj.Name()) {
			log.Printf("invalid card name: %v", fileObj.Name())
		}
	}
}
