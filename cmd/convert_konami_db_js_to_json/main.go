package main

import (
	"bytes"
	"log"
	"os"
)

func main() {
	// the inputFile is copied from
	inputFile := `C:\Users\tungd\go\src\github.com\daominah\daominah.github.io\konami_data\konami_db_en.js`
	outputFile := "konami_db.json"

	jsContent, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("error os.ReadFile %v: %v", inputFile, err)
	}

	// find the assignment '= [' to locate the JSON array start,
	// fragile but work
	marker := []byte("= [")
	idx := bytes.Index(jsContent, marker)
	if idx < 0 {
		log.Fatalf("error no '= [' found in %v", inputFile)
	}
	jsonContent := jsContent[idx+len(marker)-1:] // keep the '['

	if err := os.WriteFile(outputFile, jsonContent, 0644); err != nil {
		log.Fatalf("error os.WriteFile %v: %v", outputFile, err)
	}
	log.Printf("saved %v -> %v (%v bytes)", inputFile, outputFile, len(jsonContent))
}
