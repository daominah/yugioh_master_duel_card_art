package yugioh_master_duel_card_art

import (
	_ "embed"
	"encoding/json"
	"log"
	"strings"
)

var AllowChars = make(map[rune]bool)

func init() {
	// diff range string vs range []rune: https://stackoverflow.com/a/49062341/4097963,
	// TLDR: if don't care about rune index, same results
	for _, char := range "abcdefghijklmnopqrstuvwxyz_0123456789" {
		AllowChars[char] = true
	}
}

// NormalizeName keeps alphanumeric chars, replace others with underscore _
func NormalizeName(s string) string {
	var ret []rune
	for _, char := range strings.ToLower(s) {
		if !AllowChars[char] {
			ret = append(ret, '_')
		} else {
			ret = append(ret, char)
		}
	}
	return string(ret)
}

//go:embed konami_db.json
var allCardDataKonami []byte

/*
CardKonami is data crawled from Konami "db.yugioh-card.com", e.g.:

	{
		"CardName": "Blue-Eyes White Dragon",
		"CardType": "Monster",
		"CardSubtype": "MonsterNormal",
		"CardEffect": "This legendary dragon is a powerful engine of destruction. Virtually invincible, very few have faced this awesome creature and lived to tell the tale.",
		"CardArt": "",
		"MonsterAttribute": "LIGHT",
		"MonsterType": "Dragon",
		"MonsterLevelRankLink": 8,
		"MonsterATK": 3000,
		"MonsterATKStr": "3000",
		"MonsterDEF": 2500,
		"MonsterDEFStr": "2500",
		"MonsterAbilities": null,
		"MonsterLinkArrows": null,
		"IsNonEffectMonster": true,
		"IsPendulum": false,
		"PendulumScale": 0,
		"PendulumEffect": "",
		"MiscKonamiSet": "LOB-001",
		"MiscKonamiCardID": "4007",
		"MiscYear": "2002",
		"MiscCreator": ""
	}
*/
type CardKonami struct {
	// Konami cardID, same as file name in Master Duel data, equal to old Card.Cid
	CardID     string `json:"MiscKonamiCardID"`
	CardName   string
	MonsterATK float64
	AltArtID   string
}

func ReadAllCardDataKonami() map[string]CardKonami {
	var cardList []CardKonami
	err := json.Unmarshal(allCardDataKonami, &cardList)
	if err != nil {
		log.Fatalf("error json.Unmarshal: %v", err)
	}
	cards := make(map[string]CardKonami)
	for _, v := range cardList {
		cards[v.CardID] = v
	}

	// Create card entries for alternate art versions.
	// For each original card that has alt arts, create duplicate entries
	// keyed by the alt art IDs, with AltArtID field set to the alt art ID.
	// Example: Blue-Eyes White Dragon (4007) has alt art 3801, so we create
	// a card entry with ID "3801" containing the same data as "4007" but with AltArtID="3801".
	for origin, entry := range AltArts {
		originalCard, exists := cards[origin]
		if !exists {
			continue
		}
		for _, alt := range entry.AltArtIDs {
			v := originalCard
			v.AltArtID = alt
			cards[alt] = v
		}
	}

	log.Printf("ok ReadAllCardDataKonami, len(cards): %v", len(cards))
	return cards
}
