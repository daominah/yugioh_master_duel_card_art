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

//go:embed yugioh_card_info.json
var allCardData []byte // DEPRECATED

//go:embed konami_db_en.json
var allCardDataKonami []byte

// ReadAllCardData returns map cardId to card info,
// this func can panic if data is unexpected,
// DEPRECATED
func ReadAllCardData() map[string]Card {
	cards := make(map[string]Card)
	err := json.Unmarshal(allCardDataKonami, &cards)
	if err != nil {
		log.Fatalf("error json.Unmarshal: %v", err)
	}

	// alternative art

	cards["3415"] = Card{
		Cid:    3415,
		EnName: "I:P Masquerena",
	}
	cards["3421"] = Card{
		Cid:    3421,
		EnName: "Knightmare Unicorn",
	}
	cards["3423"] = Card{
		Cid:    3423,
		EnName: "Eldlich the Golden Lord",
	}
	cards["3863"] = Card{
		Cid:    3863,
		EnName: "Dark Magician",
	}

	// huh

	cards["15068"] = Card{
		Cid:    15068,
		EnName: "Familiar of the Evil Eye Token",
	}
	//cards["17157"] = Card{
	//	Cid:    17157,
	//	Id:     59242457,
	//	EnName: "Exosisters Magnifica",
	//}
	//cards["17344"] = Card{
	//	Cid:    17344,
	//	Id:     49858495,
	//	EnName: "Power Pro Lady Sisters",
	//}

	log.Printf("len(cards): %v", len(cards))
	if len(cards) < 1000 {
		log.Fatalf("error unexpected number of cards: %v", len(cards))
	}
	if got, want := cards["4007"].EnName, "Blue-Eyes White Dragon"; got != want {
		log.Fatalf("error unexpected card English name: %v, want: %v", got, want)
	}
	if got, want := NormalizeName(cards["4007"].EnName), "blue_eyes_white_dragon"; got != want {
		log.Fatalf("error NormalizeName: %v, want: %v", got, want)
	}
	log.Printf("loaded all cards data")
	return cards
}

// Card struct is
// DEPRECATED
/*
"4007": {
	"cid": 4007,
	"id": 89631139,
	"cn_name": "青眼白龙",
	"sc_name": "青眼白龙",
	"md_name": "青眼白龙",
	"nwbbs_n": "青眼白龙",
	"cnocg_n": "蓝眼白龙",
	"jp_ruby": "ブルーアイズ・ホワイト・ドラゴン",
	"jp_name": "青眼の白龍",
	"en_name": "Blue-Eyes White Dragon",
	"text": {
		"types": "[怪兽|通常] 龙/光\n[★8] 3000/2500",
		"pdesc": "",
		"desc": "以高攻击力著称的传说之龙。任何对手都能粉碎，其破坏力不可估量。"
	},
	"data": {
		"ot": 11,
		"setcode": 221,
		"type": 17,
		"atk": 3000,
		"def": 2500,
		"level": 8,
		"race": 8192,
		"attribute": 16
	}
}
*/
type Card struct {
	Cid    int    // Konami cardID, same as file name in Master Duel data
	Id     int    // Konami card 8-digits password
	EnName string `json:"en_name"`
	WikiEn string `json:"wiki_en"`
}

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

	// alternative art:

	cards["3415"] = cards["14676"] // I:P Masquerena
	cards["3421"] = cards["13601"] // Knightmare Unicorn
	cards["3423"] = cards["15123"] // Eldlich the Golden Lord
	cards["3863"] = cards["4041"]  // "Dark Magician"

	log.Printf("ok ReadAllCardDataKonami, len(cards): %v", len(cards))
	return cards
}
