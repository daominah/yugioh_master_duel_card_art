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
var allCardData []byte

// ReadAllCardData returns map cardId to card info,
// this func can panic if data is unexpected
func ReadAllCardData() map[string]Card {
	cards := make(map[string]Card)
	err := json.Unmarshal(allCardData, &cards)
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

/*
	{
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
	    },
	    "4083": {
	        "cid": 4083,
	        "id": 2906250,
	        "cn_name": "捕猎蛇",
	        "md_name": "捕猎蛇",
	        "nwbbs_n": "捕猎蛇",
	        "cnocg_n": "捕猎蛇",
	        "jp_ruby": "グラップラー",
	        "jp_name": "グラップラー",
	        "wiki_en": "Grappler",
	        "text": {
	            "types": "[怪兽|通常] 爬虫类/水\n[★4] 1300/1200",
	            "pdesc": "",
	            "desc": "狡猾的蛇。用又粗又长的身体勒紧对手的攻击必须注意！"
	        },
	        "data": {
	            "ot": 1,
	            "setcode": 0,
	            "type": 17,
	            "atk": 1300,
	            "def": 1200,
	            "level": 4,
	            "race": 524288,
	            "attribute": 2
	        }
	    }
	}
*/
type Card struct {
	Cid    int    // file name in Master Duel data
	Id     int    // Konami card id
	EnName string `json:"en_name"`
	WikiEn string `json:"wiki_en"`
}
