# How to Get Correct Card Names

Video transcripts garble card names (auto-captions mishear them,
and speakers often drop archetype prefixes or use nicknames).
For each video, correct every card name to its canonical English name
using the card database, so the notes stay searchable and accurate.

## The database

Canonical names live in a SQLite database in the sibling repo:

- File: `../../yugioh_card_editor_database/data/yugioh.db`
- Table `cards`: column `card_name_en` (the canonical English name).
- Table `card_texts`: `effect` per `lang` (use `lang = "en"`),
  for disambiguating by effect text.

If the sibling repo is not on this machine, fall back to the public repo
<https://github.com/daominah/yugioh_card_editor_database>.
Download only the database file (about 40 MB, plain git, not Git LFS)
into a temporary directory, never into this repo:

```bash
curl -L -o "$TEMP/yugioh.db" \
  https://raw.githubusercontent.com/daominah/yugioh_card_editor_database/main/data/yugioh.db
```

The public copy is the last pushed version,
so it can miss cards the user added locally but has not pushed yet.

Keep `yugioh_card_editor_database` unchanged: it is a repo we use as-is,
and the user may be working in it at the same time.
**Do not create files inside it**, not even a temporary command:
an untracked directory shows in the user's `git status`
and gets compiled by their `go build ./...`.

## Look up by name substring

Save a lookup script in a temporary directory outside both repos
(for example `%TEMP%\yt_transcript_scripts\lookup.py`).
It reads the live `data/yugioh.db` in place, opened read-only (`mode=ro`),
so lookups always see the user's current database and can never write to it.
It needs only the Python standard library.

```python
import sqlite3
import sys


# Absolute path to the card database in the sibling repo,
# or to the downloaded copy when the sibling repo is missing.
DATABASE = "C:/Users/tungd/go/src/github.com/daominah/yugioh_card_editor_database/data/yugioh.db"

connection = sqlite3.connect(f"file:{DATABASE}?mode=ro", uri=True)
for query in sys.argv[1:]:
    rows = connection.execute(
        "SELECT card_name_en FROM cards"
        " WHERE card_name_en LIKE '%' || ? || '%' COLLATE NOCASE"
        " ORDER BY card_name_en",
        (query,),
    )
    print(f"--- {query!r} ---")
    for (name,) in rows:
        print(f"  {name}")
```

Pass each guessed name (or archetype prefix) as an argument:

```bash
python "$TEMP/yt_transcript_scripts/lookup.py" "Vanquish Soul" "Lunalight" "Dominus"
```

To disambiguate by effect, query `card_texts` the same way
(join on the card id, filter `lang = 'en'`, match on `effect`).

For the full details of one card (stats, all locales, effect, prints),
run the existing command with a `card_id` from the `yugioh_card_editor_database` directory
(`go run` only reads the repo, it creates no files there):

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

## In the notes file

- Use the full canonical name on first mention in each major section
  (for example Summary, Condensed, Transcript), then the short form is fine.
- In the Transcript preamble, note that names were corrected against the database
  and list the non-obvious fixes (garbled to canonical).
