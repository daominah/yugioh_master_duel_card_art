package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//go:embed card_info.json
var allCardData []byte

var (
	dirSourceCardArt string
	dirTargetCardArt string
)

func init() {
	flag.StringVar(&dirSourceCardArt, "dirSourceCardArt",
		`/media/tungdt/WindowsData/picture/card_Master_Duel_art`,
		"path to source directory that contains extracted images from the game")
	flag.StringVar(&dirTargetCardArt, "dirTargetCardArt",
		`/home/tungdt/opt/card_Master_Duel/in_game_art_renamed`,
		"path to target directory that contains output renamed images from this program")
}

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

	flag.Parse()
	log.Printf("flag vars:")
	log.Printf("dirSourceCardArt: %v", dirSourceCardArt)
	log.Printf("dirTargetCardArt: %v", dirTargetCardArt)

	cards := make(map[string]Card)
	err := json.Unmarshal(allCardData, &cards)
	if err != nil {
		log.Fatalf("error json.Unmarshal: %v", err)
	}

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

	log.Printf("doing read dir %v", dirSourceCardArt)
	dir, err := os.ReadDir(dirSourceCardArt)
	if err != nil {
		log.Fatalf("error os.ReadDir: %v", err)
	}

	nCopiedFiles := 0
	for i, f := range dir {
		//if i > 100 { // small number for testing
		//	break
		//}
		sourceFullPath := filepath.Join(dirSourceCardArt, f.Name())
		fnameWOExt := strings.TrimSuffix(f.Name(), ".png")
		cardInfo, found := cards[fnameWOExt]
		if !found {
			log.Printf("i %v ignore %v", i, f.Name())
			continue
		}
		enName := cardInfo.EnName
		if enName == "" {
			enName = cardInfo.WikiEn
		}
		targetName := fmt.Sprintf("%v_%v_%v.png", NormalizeName(enName), cardInfo.Id, cardInfo.Cid)
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
		nCopiedBytes, err := io.Copy(targetFile, sourceFile)
		if err != nil {
			log.Printf("error io.Copy: %v", err)
			continue
		}
		nCopiedFiles += 1
		log.Printf("created new file %v nCopiedBytes %v", targetName, nCopiedBytes)
	}
	log.Printf("func main returned, nCopiedFiles: %v", nCopiedFiles)
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
