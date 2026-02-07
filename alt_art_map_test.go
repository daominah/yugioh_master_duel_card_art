package yugioh_master_duel_card_art

import (
	"os"
	"testing"
)

func TestInitAltArtsMap(t *testing.T) {
	if len(AltArts) < 1 {
		t.Fatalf("error AltArts map is empty after initialization")
	}
	t.Logf("AltArts: %+v", AltArts)

	if len(AltArts["4041"].AltArtIDs) < 2 {
		t.Fatalf("error AltArts incomplete, expected Dark Magician has 2nd alt art on 4th anniversary event")
	}
}

func TestSaveAltArtsJSONAsJavascript(t *testing.T) {
	outputFile := "alt_arts.js"
	f, err := os.Create(outputFile)
	if err != nil {
		t.Fatalf("error os.Create %v: %v", outputFile, err)
	}
	f.WriteString(`const AltArts = `)
	f.Write(altArtsJSON)
	f.Close()
	t.Logf("saved alt arts JSON as JavaScript file: %v", outputFile)
}
