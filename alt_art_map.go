package yugioh_master_duel_card_art

import (
	_ "embed"
	"encoding/json"
	"log"
	"strconv"
)

// List of all alternative arts in Master Duel can be found here:
// https://yugipedia.com/wiki/Category:Yu-Gi-Oh!_Master_Duel_cards_with_alternate_artworks

//go:embed alt_arts.json
var altArtsJSON []byte

// AltArts maps original card ID to AltArtEntry.
// One original card can have multiple alternative arts.
// This map is loaded from alt_arts.json at package initialization.
var AltArts map[string]AltArtEntry

func init() {
	var entries []AltArtEntry
	err := json.Unmarshal(altArtsJSON, &entries)
	if err != nil {
		log.Fatalf("error loading alt_arts.json: %v", err)
	}

	AltArts = make(map[string]AltArtEntry)
	for _, entry := range entries {
		// validate OriginalCardID and all AltArtIDs in JSON data are valid integers
		_, err := strconv.Atoi(entry.OriginalCardID)
		if err != nil {
			log.Fatalf("error invalid OriginalCardID in alt_arts.json: %q is not a valid integer: %v", entry.OriginalCardID, err)
		}
		for _, altID := range entry.AltArtIDs {
			_, err := strconv.Atoi(altID)
			if err != nil {
				log.Fatalf("error invalid AltArtID in alt_arts.json: OriginalCardID %q has invalid AltArtID %q: %v", entry.OriginalCardID, altID, err)
			}
		}

		AltArts[entry.OriginalCardID] = entry
	}
	log.Printf("loaded AltArts from JSON, len: %v", len(AltArts))
}

// AltArtEntry represents an alt art entry in the JSON file.
type AltArtEntry struct {
	OriginalCardID string   `json:"OriginalCardID"`
	CardName       string   `json:"CardName"`
	AltArtIDs      []string `json:"AltArtIDs"`
}
