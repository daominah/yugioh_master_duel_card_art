package main

import (
	"fmt"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	yugioh "github.com/daominah/yugioh_master_duel_card_art"
)

// this program collects the main art layer of every Master Duel profile wallpaper
// (extracted from YuGiOh Master Duel, see README.md for more info),
// then copies it to "${dirTarget}" named
// "{number}_{layer}_{name}_{id}[_alt{altID}]_{region}.png", e.g.
// "0001_1_blue_eyes_alternative_white_dragon_12253_tcg.png".
// The name, id and "_alt" part follow cmd/rename_from_extracted, so a wallpaper
// carries the same label as the card art it was cut from.
//
// A wallpaper is not shipped as one finished image: the game composites it at
// runtime from a transparent art layer, the shared effect textures in
// "wallpaper/common/", and a background gradient. The art layer
// "WallPaper{number}_1.png" is the piece worth collecting: it holds the
// foreground subject, cut out of its background.
//
// 121 of the 137 wallpapers that ship numbered layers also have a "_2" (and
// sometimes "_3"), almost always the background that "_1" was cut from, which is
// why "_1" alone is enough. The four that invert this are listed in
// wallpaperArtFiles, so the layer in the name is read from the source rather
// than assumed to be "_1". That map also covers the 138th wallpaper, 1001, whose
// art is not a numbered layer at all.
//
// The layer has no fixed size: it ranges from 200x200 up to 2048x1304,
// so the biggest variant is picked by pixel area rather than by an expected size.

// Inputs (edit for each run):
const (
	// dirSourceWallpaper is the extracted wallpaper tree, holding one
	// "wallpaper{number}/{region}/" folder per wallpaper.
	dirSourceWallpaper = `D:\tmp_process_MD_file_by_path\assets\resourcesassetbundle\wallpaper`

	// dirTarget must exist before running, this program does not create it.
	dirTarget = `D:\tmp_process_MD_file\MD_wallpaper`
)

// wallpaperCard is the card a wallpaper took its art from.
// The zero value means no card is known for that wallpaper.
type wallpaperCard struct {
	// name is the card name as Konami writes it, e.g. "Sky Striker Ace - Kagari".
	// It is stored unnormalized so it can be checked against konami_db.json by eye,
	// yugioh.NormalizeName turns it into the file name form at use.
	// It is empty when the art matched a card id that no name source knows yet.
	name string

	// id is the Konami card ID, the same ID cmd/rename_from_extracted puts in
	// its file names.
	id string

	// altID is set when the wallpaper uses an alternative art of id rather
	// than the original art. Master Duel gives each alternative art its own ID,
	// which is the ID that matched, but the card it belongs to is id.
	altID string
}

// wallpaperCards maps a wallpaper number to the card its art comes from.
//
// Most wallpapers are not original art: they are the illustration of one card,
// cut out of its background. So the card here is the one that owns the
// illustration, which is not always a monster: 43 of these are Spell or Trap
// cards whose art happens to depict a monster or a scene. Wallpaper 0052 is the
// art of the trap "Sangen Kaiho", which pictures Five-Headed Dragon, and is
// named for the trap.
//
// Most entries were derived by matching each wallpaper against the extracted
// card art (see cmd/mix_dir_card_id), then reading the name out of
// konami_db.json and alt_arts.json. Matching is why they can be trusted: the
// illustration lines up with the wallpaper point for point.
//
// The rest were named by hand, because matching alone could not name them:
//   - the card art is not in the extracted dump (0023, 0024), so there was
//     nothing to match against.
//   - the art is a token (0127), which no name source lists.
//   - layer _1 holds only part of the illustration (0071), too little to match.
//   - the art matched, but no name source knew the id yet (0108).
//   - the art is not in layer _1 at all, so the matcher never saw it
//     (0015, 1001), see wallpaperArtFiles.
//
// An entry with a name but no id is named for what the art shows rather than
// for a card it was cut from, because no single card owns it: the WCS2025
// gallery (0086) paints five cards at once, the anniversary pieces (0124) two,
// and the collaborations (2002) none. id stays empty there so that it always
// means one thing: the card whose art this is.
//
// An empty entry is a wallpaper still to be named. The run warns about those and
// about the ones left without a id, so the list stays visible.
var wallpaperCards = map[string]wallpaperCard{
	"0001": {"Blue-Eyes Alternative White Dragon", "12253", ""},
	"0002": {"El Shaddoll Construct", "11258", ""},
	"0003": {"Tri-Brigade Shuraig the Ominous Omen", "15527", ""},
	"0004": {"Dingirsu, the Orcust of the Evening Star", "14288", ""},
	"0005": {"Sky Striker Ace - Kagari", "13668", ""},
	"0006": {"Eldlich the Golden Lord", "15123", ""},
	"0007": {"Mekk-Knight Crusadia Avramax", "14297", ""},
	"0008": {"Meteonis Drytron", "15642", ""},
	"0009": {"Palladium Oracle Mahad", "12357", ""},
	"0010": {"Elemental HERO Honest Neos", "12901", ""},
	"0011": {"Shooting Quasar Dragon", "9709", ""},
	"0012": {"Number F0: Utopic Draco Future", "14958", ""},
	"0013": {"Odd-Eyes Arc Pendulum Dragon", "13359", ""},
	"0014": {"Accesscode Talker", "15032", ""},
	"0015": {"Ghostrick Festival", "16852", ""},
	"0016": {"The World Legacy", "14688", ""},
	"0017": {"Dark Magician", "4041", "3863"},
	"0018": {"Vernusylph of the Flourishing Hills", "17418", ""},
	"0019": {"Ursarctic Radiation", "16864", ""},
	"0020": {"Abyss Actors' Curtain Call", "13866", ""},
	"0021": {"Rescue-ACE Turbulence", "17997", ""},
	"0022": {"Master Peace, the True Dracoslayer", "12419", ""},
	// both arts are missing from the extracted card dump, so there was nothing
	// for the matcher to compare against
	"0023": {"Legendary Magician of Dark", "10322", ""},
	"0024": {"Legendary Dragon of White", "10321", ""},
	"0025": {"Vanquish Soul Caesar Valius", "18732", ""},
	"0026": {"Branded Fusion", "17066", ""},
	"0027": {"New Frontier", "18517", ""},
	"0028": {"Gold Pride - Star Leon", "18458", ""},
	"0029": {"Gift Exchange", "11661", ""},
	"0030": {"Onikuji", "12397", ""},
	"0031": {"Loka Samsara", "19215", ""},
	"0032": {"Blue-Eyes White Dragon", "4007", "3801"},
	"0033": {"Magicians of Bonds and Unity", "18868", ""},
	"0034": {"Dragon of Pride and Soul", "20260", ""},
	"0035": {"Promethean Princess, Bestower of Flames", "19507", ""},
	"0036": {"Chimera the Illusion Beast", "18821", ""},
	"0037": {"Mamemaki", "13633", ""},
	"0038": {"Red Dragon Archfiend", "7735", ""},
	"0039": {"Snake-Eyes Flamberge Dragon", "19152", ""},
	"0040": {"Flawless Perfection of the Tenyi", "14503", ""},
	"0041": {"Number 41: Bagooska the Terribly Tired Tapir", "13163", ""},
	"0042": {"Sky Striker Mobilize - Linkage!", "17034", ""},
	"0043": {"Stars Align across the Milky Way", "17834", ""},
	"0044": {"Sacred Fire King Garunix", "19397", ""},
	"0045": {"Dark Hole Dragon", "19162", ""},
	"0046": {"Kingyo Sukui", "14390", ""},
	"0047": {"Galaxy-Eyes Photon Dragon", "9729", ""},
	"0048": {"Raidraptor - Rising Rebellion Falcon", "19500", ""},
	"0049": {"Ancient Gear Advance", "19896", ""},
	"0050": {"Aroma Healing", "19532", ""},
	"0051": {"Final Gesture", "10285", ""},
	"0052": {"Sangen Kaiho", "20252", ""},
	"0053": {"Stardust Dragon", "7734", "3882"},
	"0054": {"Junk Warrior", "7696", "19077"},
	"0055": {"Goblin Circus", "10426", ""},
	"0056": {"Ties That Bind", "19894", ""},
	"0057": {"Terrors of the Afterroot", "19534", ""},
	"0058": {"Susurrus of the Sinful Spoils", "20239", ""},
	"0059": {"Varudras, the Final Bringer of the End Times", "19886", ""},
	"0060": {"The Unstoppable Exodia Incarnate", "20212", ""},
	"0061": {"Sky Striker Ace - Raye", "13670", "21227"},
	"0062": {"Reinforcement of the Army", "5328", "20040"},
	"0063": {"Dogmatika Encounter", "15305", ""},
	"0064": {"Majesty of the White Dragons", "20606", ""},
	"0065": {"Varar, Vaalmonican Concord", "20228", ""},
	"0066": {"Dominus Purge", "20257", ""},
	"0067": {"Dominus Impulse", "20555", ""},
	"0068": {"Primite Howl", "20552", ""},
	"0069": {"Goblin Bikers Gone Wild", "20488", ""},
	"0070": {"Elemental HERO Neos", "6653", "3881"},
	// the art is split over three layers, so layer _1 alone is only the left
	// figure: not enough for the matcher to place it, confirmed by eye instead
	"0071": {"Sky Striker Maneuver - Scissors Cross", "14871", ""},
	"0072": {"Traptrix Sera", "14203", ""},
	"0073": {"Argostars - Glorious Adra", "20749", ""},
	"0074": {"Firestorms Over Atlantis", "20542", ""},
	// a fresh monochrome illustration of the characters rather than that card's
	// art, so it is named for what it shows and keeps no id
	"0075": {"Story of White monochrome The Fallen & The Virtuous", "", ""},
	"0076": {"Maliss in Underground", "20588", ""},
	"0077": {"Liberator Eto", "20762", ""},
	"0078": {"Sunset Beat", "19831", ""},
	"0079": {"Tenyi Spirit - Suruya", "20760", ""},
	"0080": {"Shateki", "21241", ""},
	"0081": {"Aluber the Jester of Despia", "16195", "20569"},
	"0082": {"Blazing Cartesia, the Virtuous", "17767", "20570"},
	"0083": {"Medallion of the Ice Barrier", "9131", "19736"},
	"0084": {"Black Rose Dragon", "7898", "3420"},
	"0085": {"A Bao A Qu, the Lightless Shadow", "20786", ""},
	// layer _1 is only the Millennium Puzzle emblem, the Exodia art is the
	// gallery in layer _2: Exodia the Forbidden One framed between its four limbs
	"0086": {"WCS2025 celebratory Exodia", "", ""},
	"0088": {"Invoked Magistus Omega", "21577", ""},
	"0089": {"Allied Code Talker @Ignister", "21188", ""},
	"0090": {"Lacrima the Crimson Tears", "20490", ""},
	"0091": {"Galatea-I, the Orcust Automaton", "20947", ""},
	"0092": {"Dragonmaid Hospitality", "14766", "3436"},
	"0094": {"Lollipo★Yummy Way", "21370", ""},
	"0095": {"Diabellstar the Black Witch", "19148", "21231"},
	"0096": {"Lady Labrynth of the Silver Castle", "17785", "21230"},
	"0097": {"Archfiend's Ghastly Glitch", "17372", ""},
	"0098": {"Eldlich the Golden Lord", "15123", "3423"},
	"0099": {`K9-X "Werewolf"`, "21382", ""},
	"0100": {"Ohime the Manifested Mikanko", "18017", "21733"},
	"0101": {"Mikanko Shinbu - Noble Twins", "21617", ""},
	"0102": {"Combined Maneuver - Engage Zero!", "19933", ""},
	"0103": {"Ketu Dracotail", "21361", ""},
	"0104": {"First of the Dragonlords", "21447", ""},
	"0105": {"Diabellze the Original Sinkeeper", "19853", ""},
	"0106": {"Megalith Anastasis", "21832", ""},
	"0107": {"Solfachord Symphony", "18528", ""},
	// too new for konami_db.json, the art matched id 22715 but nothing named it
	"0108": {"W:P Fancy Ball", "22715", ""},
	"0109": {"I:P Masquerena", "14676", "22746"},
	"0110": {"S:P Little Knight", "19188", "22747"},
	"0111": {"I:P Masquerena", "14676", "3415"},
	"0112": {"S:P Little Knight", "19188", ""},
	"0113": {"Spright Gamma Burst", "17458", ""},
	"0114": {"Nightwinged Cleric", "22141", ""},
	"0115": {"Radiant Typhoon Fonix, the Great Flame", "21783", ""},
	"0116": {"Destiny HERO - Destroyer Phoenix Enforcer", "16524", ""},
	"0117": {"Kewl Tune B2B", "22551", ""},
	"0118": {"Jupiter the Power Patron of Destruction", "21810", ""},
	"0119": {"Dominus Spiral", "21845", ""},
	"0120": {"Masked HERO Dark Law", "11313", "21730"},
	"0122": {"True King of All Calamities", "12960", ""},
	"0123": {"Dark Magician", "4041", "3880"},
	// an anniversary piece drawing two cards at once, so no single card owns it
	"0124": {"4th Anniversary anime I:P Masquerena & S:P Little Knight", "", ""},
	"0125": {"Elfnotes: Aristeia of Trust", "22174", ""},
	"0126": {"Elfnotes: Welcome Home", "22159", ""},
	// a token, so the id is in neither konami_db.json nor alt_arts.json
	"0127": {"The Virtuous Vestals", "17075", ""},
	"0128": {"Ryzeal Mass Driver", "21212", ""},
	"0129": {"Hecahands Bait", "22584", ""},
	"0130": {"Junora the Power Patron of Tuning", "22142", ""},
	"0131": {"Triple Tactics Thrust", "18214", "22189"},
	"0132": {"Fusion Deployment", "15057", "22188"},
	"0133": {"Weaver of Fairy Tails", "22542", ""},
	"0135": {"Gorgon of Zilofthonia", "21461", ""},
	"0136": {"Fydraulis Harmonia", "22533", ""},
	"1001": {"Destined Rivals", "15579", ""},
	"2001": {"Rescue Rabbit", "9755", "21965"},
	// a collaboration piece, not a card art at all
	"2002": {"eFootball collab Neymar", "", ""},
	"2003": {"That's 10!", "20246", ""},
	"2004": {"Endymion, the Mighty Master of Magic", "14439", ""},
	"2005": {"Emerging Emergency Rescute Rescue", "13100", ""},
}

// wallpaperArtFiles overrides which source asset holds the art, for the few
// wallpapers where it is not the usual "WallPaper{number}_1.png".
//
// Layer "_1" normally holds the foreground subject and "_2" the background it
// was cut from, so "_1" alone is the art. These four invert that, and taking
// "_1" would collect the wrong image rather than no image, which is why they are
// listed by hand instead of guessed at:
//   - 0015 keeps the Ghostrick Festival scene in "_2" and cuts the foreground
//     into "_1_1".."_1_3" plus a sprite atlas, so "_1" does not exist.
//   - 0086 keeps only the Millennium Puzzle emblem in "_1", the Exodia gallery
//     is "_2".
//   - 0133 keeps only the ice sprites in "_1", the scene they belong to is "_2".
//   - 1001 is not a numbered layer at all: its art is the World Championship
//     title background, which is why a layer number could not name it.
//
// The value is matched the same way as the default layer, so the "_#{pathID}"
// extraction variants of it are still picked up, biggest one winning.
var wallpaperArtFiles = map[string]string{
	"0015": "WallPaper0015_2.png",
	"0086": "WallPaper0086_2.png",
	"0133": "WallPaper0133_2.png",
	"1001": "Title_WCSBg.png",
}

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

	if _, err := os.Stat(dirTarget); err != nil {
		log.Fatalf("error probably directory does not exist: %v: %v", dirTarget, err)
		// mkdir MD_wallpaper
	}
	log.Printf("target directory exists, READY to process")

	beginProcessT := time.Now()
	wallpaperDirs, err := os.ReadDir(dirSourceWallpaper)
	if err != nil {
		log.Fatalf("error os.ReadDir: %v", err)
	}

	nWallpapers := 0
	nCopiedFiles := 0
	var unnamedNumbers []string    // wallpapers missing from wallpaperCards
	var skippedWallpapers []string // wallpapers shipping no main art layer
	for _, wallpaperDir := range wallpaperDirs {
		// skip "common" (shared effect textures) and "thumbui" (menu chrome),
		// they hold no per-wallpaper art
		number := strings.TrimPrefix(wallpaperDir.Name(), "wallpaper")
		if !wallpaperDir.IsDir() || number == wallpaperDir.Name() {
			continue
		}
		nWallpapers += 1

		// an unknown wallpaper and one left blank in the map are the same case:
		// collect the art, but warn that it lands without a card name
		card := wallpaperCards[number]

		// a wallpaper ships its art once per region, and the two regions are not
		// always the same art, so both are collected and the region goes in the
		// file name to keep them apart
		wallpaperPath := filepath.Join(dirSourceWallpaper, wallpaperDir.Name())
		regionDirs, err := os.ReadDir(wallpaperPath)
		if err != nil {
			log.Printf("error os.ReadDir: %v", err)
			continue
		}
		isCollected := false
		for _, regionDir := range regionDirs {
			// loose files sitting next to the region folders are duplicates of
			// assets that carry no region in their container path
			if !regionDir.IsDir() {
				continue
			}
			region := regionDir.Name()
			sourceFullPath := findMainArtLayer(filepath.Join(wallpaperPath, region), number)
			if sourceFullPath == "" {
				continue
			}
			isCollected = true

			// the "WallPaper" prefix is dropped: every file here is one, so it only
			// pushes the number away from the start of the name. The layer suffix
			// stays, it is part of the source asset name, and it is read from the
			// source rather than assumed to be "_1", so the few wallpapers whose
			// art lives elsewhere say so.
			//
			// The source may be a "_#{pathID}" variant, but the target never keeps
			// that suffix: it is an artifact of two assets sharing a container name,
			// and means nothing outside the extraction.
			// the name, id and region follow cmd/rename_from_extracted, so a
			// wallpaper sorts next to the card art it was cut from
			targetName := fmt.Sprintf("%v_%v", number,
				artLayerLabel(filepath.Base(sourceFullPath), number))
			if card.name != "" {
				targetName += "_" + yugioh.NormalizeName(card.name)
			}
			if card.id != "" {
				targetName += "_" + card.id
			}
			if card.altID != "" {
				targetName += "_alt" + card.altID
			}
			targetName += "_" + region + ".png"

			// copyFile does nothing if target file existed
			isCopied := copyFile(sourceFullPath, filepath.Join(dirTarget, targetName))
			if isCopied {
				nCopiedFiles += 1
				log.Printf("created new file %v (from %v)", targetName, filepath.Base(sourceFullPath))
			}
		}
		// a skipped wallpaper is not also reported as unnamed:
		// nothing landed in dirTarget for it, so its name never mattered
		if !isCollected {
			skippedWallpapers = append(skippedWallpapers, number)
		} else if card.name == "" {
			unnamedNumbers = append(unnamedNumbers, number)
		}
	}

	log.Printf("-------------------------------------------------------")
	log.Printf("func main returned")
	log.Printf("done processing %v wallpapers, duration: %v", nWallpapers, time.Since(beginProcessT))
	log.Printf("nCopiedFiles: %v", nCopiedFiles)
	// a wallpaper with no "_1" layer and no wallpaperArtFiles override: look at
	// its layers by hand and add the one holding the art to that map
	if len(skippedWallpapers) > 0 {
		log.Printf("WARNING no main art layer, skipped: %v", strings.Join(skippedWallpapers, ", "))
	}
	// an unnamed wallpaper is collected as bare "{number}_{layer}_{region}.png",
	// so it is still usable, it just carries no hint of what it shows
	if len(unnamedNumbers) > 0 {
		sort.Strings(unnamedNumbers)
		log.Printf("WARNING no name in wallpaperCards, collected with the number only: %v",
			strings.Join(unnamedNumbers, ", "))
	}
}

// pathIDSuffixRegexp matches the "_#{pathID}" suffix AssetStudio adds when two
// assets share a container name. It means nothing outside the extraction.
var pathIDSuffixRegexp = regexp.MustCompile(`_#\d+$`)

// artLayerRegexp matches the art layer to collect for wallpaper "{number}":
// "WallPaper0001_1.png" by default, or the wallpaperArtFiles override.
// Either way it also matches that asset's "_#{pathID}" extraction variants,
// e.g. "WallPaper0001_1_#556.png".
//
// The default deliberately rejects the "_1_1", "_1_2", ... split sub-layers and
// the "_2" background layer, neither of which is the art on its own.
func artLayerRegexp(number string) *regexp.Regexp {
	artFile, isException := wallpaperArtFiles[number]
	if !isException {
		artFile = fmt.Sprintf("WallPaper%v_1.png", number)
	}
	stem := strings.TrimSuffix(artFile, ".png")
	return regexp.MustCompile(`^` + regexp.QuoteMeta(stem) + `(_#\d+)?\.png$`)
}

// findMainArtLayer returns the path of the biggest art layer in regionPath,
// or "" if the wallpaper ships none.
// Several variants of the same layer can exist, differing slightly in size
// (e.g. 2048x1304 vs 2018x1302), so the one with the most pixels wins.
func findMainArtLayer(regionPath string, number string) string {
	files, err := os.ReadDir(regionPath)
	if err != nil {
		log.Printf("error os.ReadDir: %v", err)
		return ""
	}
	artLayer := artLayerRegexp(number)
	bestPath, bestPixels := "", 0
	for _, f := range files {
		if f.IsDir() || !artLayer.MatchString(f.Name()) {
			continue
		}
		fullPath := filepath.Join(regionPath, f.Name())
		width, height := readPNGSize(fullPath)
		if width*height > bestPixels {
			bestPath, bestPixels = fullPath, width*height
		}
	}
	return bestPath
}

// artLayerLabel returns the layer part of the target name, read from the source
// asset name so that the target says which layer it came from:
// "WallPaper0015_2.png" gives "2".
// An asset that is not a numbered "WallPaper{number}_{layer}" layer, like
// wallpaper 1001's "Title_WCSBg.png", is labelled with its own name instead.
func artLayerLabel(sourceName string, number string) string {
	stem := pathIDSuffixRegexp.ReplaceAllString(strings.TrimSuffix(sourceName, ".png"), "")
	layer := strings.TrimPrefix(stem, fmt.Sprintf("WallPaper%v_", number))
	if layer != stem {
		return layer
	}
	return yugioh.NormalizeName(stem)
}

// readPNGSize returns (width, height) of a PNG, or (0, 0) on error.
// It only decodes the header, so it stays cheap on the 2048px layers.
func readPNGSize(fullPath string) (int, int) {
	f, err := os.Open(fullPath)
	if err != nil {
		log.Printf("error os.Open: %v", err)
		return 0, 0
	}
	defer f.Close()
	config, err := png.DecodeConfig(f)
	if err != nil {
		log.Printf("error png.DecodeConfig %v: %v", fullPath, err)
		return 0, 0
	}
	return config.Width, config.Height
}

// copyFile does nothing if target file existed,
// copyFile logs and handles error too
func copyFile(sourceFullPath string, targetFullPath string) bool {
	sourceFile, err := os.Open(sourceFullPath)
	if err != nil {
		log.Printf("error os.Open: %v", err)
		return false
	}
	defer sourceFile.Close()
	if _, err := os.Stat(targetFullPath); err == nil {
		// log.Printf("do nothing because target file existed %v", targetFullPath)
		return false
	}
	targetFile, err := os.Create(targetFullPath)
	if err != nil {
		log.Printf("error os.Create: %v", err)
		return false
	}
	defer targetFile.Close()
	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		log.Printf("error io.Copy: %v", err)
		return false
	}
	return true
}
