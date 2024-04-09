package yugioh_master_duel_card_art

import (
	_ "embed"
	"encoding/json"
	"log"
	"strings"
)

var AllowChars = make(map[rune]bool)

func init() {
	for _, char := range []rune("abcdefghijklmnopqrstuvwxyz_0123456789") {
		AllowChars[char] = true
	}
}

func NormalizeName(s string) string {
	var ret []rune
	for _, char := range []rune(strings.ToLower(s)) {
		if !AllowChars[char] {
			ret = append(ret, '_')
		} else {
			ret = append(ret, char)
		}
	}
	return string(ret)
}

//go:embed konami_db_en.json
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

	// map alt art ID to original art ID
	altArts := map[string]string{
		"3401": "5000",  // The Winged Dragon of Ra
		"3411": "14496", // Apollousa, Bow of the Goddess
		"3415": "14676", // I:P Masquerena
		"3421": "13601", // Knightmare Unicorn
		"3423": "15123", // Eldlich the Golden Lord
		"3801": "4007",  // Blue-Eyes White Dragon
		"3863": "4041",  // Dark Magician
		"3868": "4998",  // Obelisk the Tormentor
		"3869": "4999",  //  Slifer the Sky Dragon
		"3891": "12950", // Ash Blossom & Joyous Spring
	}
	for alt, origin := range altArts {
		cards[alt] = cards[origin]
	}

	log.Printf("ok ReadAllCardDataKonami, len(cards): %v", len(cards))
	return cards
}
