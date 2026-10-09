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

## Lookup scripts

Save these scripts in a temporary directory outside both repos,
one directory per run (for example `%TEMP%\yt_transcript_scripts_2015\`),
so parallel runs don't overwrite each other's `YEAR`.
They read the live `data/yugioh.db` in place, opened read-only (`mode=ro`),
so lookups always see the user's current database and can never write to it.
They need only the Python standard library,
and they print UTF-8 so names with ★, ☆ or Ø don't crash a Windows console.

### Look up by name substring

`lookup.py`, one argument per guessed name (or archetype prefix):

```python
import sqlite3
import sys

# Absolute path to the card database in the sibling repo,
# or to the downloaded copy when the sibling repo is missing.
DATABASE = "C:/Users/tungd/go/src/github.com/daominah/yugioh_card_editor_database/data/yugioh.db"

sys.stdout.reconfigure(encoding="utf-8")
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

```bash
python "$TEMP/yt_transcript_scripts_2015/lookup.py" "Vanquish Soul" "Lunalight" "Dominus"
```

### Search by effect text

`effect_search.py`, for a garble that sounds nothing like the name:
every argument must appear in the English effect text.
Set `YEAR` to the recap year to list only that year's new cards
(`cards.year` is text, for example `"2025"`), or to `""` for all years.
With `YEAR` set and no arguments, it lists every card of that year,
the fastest way to map a year's new garbled names.

```python
import sqlite3
import sys

DATABASE = "C:/Users/tungd/go/src/github.com/daominah/yugioh_card_editor_database/data/yugioh.db"
YEAR = "2025"

sys.stdout.reconfigure(encoding="utf-8")
connection = sqlite3.connect(f"file:{DATABASE}?mode=ro", uri=True)
query = (
    "SELECT c.card_name_en, t.effect FROM cards c"
    " JOIN card_texts t ON t.card_id = c.card_id AND t.lang = 'en'"
    " WHERE (? = '' OR c.year = ?)"
)
params = [YEAR, YEAR]
for word in sys.argv[1:]:
    query += " AND t.effect LIKE '%' || ? || '%' COLLATE NOCASE"
    params.append(word)
for name, effect in connection.execute(query + " ORDER BY c.card_name_en", params):
    print(f"{name}\n  {effect[:200]}")
```

```bash
python "$TEMP/yt_transcript_scripts_2015/effect_search.py" "add 1" "then discard 1"
```

### Check every backticked name

`check_names.py`, run on each finished `.md` file:
it lists every backticked name that is not exactly a `cards.card_name_en`.
Only card names may be backticked, so the list must come out empty
(file names and code in a guide like this one are the exception).

```python
import re
import sqlite3
import sys

DATABASE = "C:/Users/tungd/go/src/github.com/daominah/yugioh_card_editor_database/data/yugioh.db"

sys.stdout.reconfigure(encoding="utf-8")
connection = sqlite3.connect(f"file:{DATABASE}?mode=ro", uri=True)
for path in sys.argv[1:]:
    with open(path, encoding="utf-8") as file:
        names = set(re.findall(r"`([^`\n]+)`", file.read()))
    failed = sorted(
        name for name in names
        if not connection.execute("SELECT 1 FROM cards WHERE card_name_en = ?", (name,)).fetchone()
    )
    print(f"{path}: {len(names)} names, {len(failed)} failed")
    for name in failed:
        print(f"  {name}")
```

```bash
python "$TEMP/yt_transcript_scripts_2015/check_names.py" year_2015/recap2015_*.md
```

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
