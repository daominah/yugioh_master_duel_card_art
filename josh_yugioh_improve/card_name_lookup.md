# How to Get Correct Card Names

Video transcripts garble card names (auto-captions mishear them,
and Josh often drops archetype prefixes when speaking).
For each episode, correct every card name to its canonical English name
using the card database, so the notes stay searchable and accurate.

## The database

Canonical names live in a SQLite database in the sibling repo:

- File: `../../yugioh_card_editor/data/yugioh.db`
- Table `cards`: column `card_name_en` (the canonical English name).
- Table `card_texts`: `effect` per `lang` (use `lang = "en"`),
  for disambiguating by effect text.

Keep `yugioh_card_editor` unchanged: it is a repo we use as-is.
Any lookup script below is temporary. Delete it after the run
so `git status` in that repo stays clean.

## Look up by name substring

Create a throwaway command in `yugioh_card_editor`, run it, then delete it.
It mirrors the existing `cmd/read-yugiohdb-by-cardid`
(same `base.GetProjectRootDir` plus `modernc.org/sqlite` setup).

```go
// cmd/search-yugiohdb-by-name/search_yugiohdb_by_name.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/daominah/yugioh_card_editor/pkg/base"
	_ "modernc.org/sqlite"
)

func main() {
	projectRoot, err := base.GetProjectRootDir()
	if err != nil {
		log.Fatalf("error base.GetProjectRootDir: %v", err)
	}
	db, err := sql.Open("sqlite", filepath.Join(projectRoot, "data/yugioh.db"))
	if err != nil {
		log.Fatalf("error sql.Open: %v", err)
	}
	defer db.Close()

	for _, query := range os.Args[1:] {
		fmt.Printf("\n--- %q ---\n", query)
		rows, err := db.Query(
			`SELECT card_name_en FROM cards
             WHERE card_name_en LIKE '%' || ? || '%' COLLATE NOCASE
             ORDER BY card_name_en`, query)
		if err != nil {
			log.Fatalf("error db.Query: %v", err)
		}
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				log.Fatalf("error rows.Scan: %v", err)
			}
			fmt.Printf("  %s\n", name)
		}
		rows.Close()
	}
}
```

Run it from the `yugioh_card_editor` directory,
passing each guessed name (or archetype prefix) as an argument:

```bash
go run ./cmd/search-yugiohdb-by-name "Vanquish Soul" "Lunalight" "Dominus"
```

Then remove the `cmd/search-yugiohdb-by-name` directory.

For the full details of one card (stats, all locales, effect, prints),
use the existing command with a `card_id`:

```bash
go run ./cmd/read-yugiohdb-by-cardid 19375
```

## Disambiguating a garbled name

The captions rarely give the exact name, so match on more than spelling:

- **Search the archetype prefix**, not the whole phrase.
  "heavy borger" is inside the `Vanquish Soul` list as `Vanquish Soul Heavy Borger`.
- **Pick the closest phonetic entry.**
  "Lucius" is `Dracotail Lukias`, "Alibur" is `Aluber the Jester of Despia`.
- **Confirm by effect** when a name doesn't match directly.
  "backup adnister" turned out to be `Backup @Ignister`,
  found by reading which Maliss-playable card has the "add 1, then discard 1" effect.
- **Match the described mechanic.**
  Josh's "liger" needs four monsters and sends one from the extra deck:
  that is `Lunalight Liger Dancer` (`Lunalight Leo Dancer` + 3), not Leo Dancer itself.

## In the episode file

- Use the full canonical name on first mention in each major section
  (Summary, Condensed, Transcript), then the short form is fine.
- In the Transcript preamble, note that names were corrected against the database
  and list the non-obvious fixes (garbled to canonical).
