# TheLawYGO TCG History Recap

Notes on [TheLawYGO](https://www.youtube.com/@TheLawYGO)'s yearly
"The {year} Yugioh TCG Recap" videos, one directory per year.

| Year | Video                                         |
|------|-----------------------------------------------|
| 2016 | <https://www.youtube.com/watch?v=y8S7Ns89Q8A> |
| 2017 | <https://www.youtube.com/watch?v=44Otn2Uo0_8> |
| 2022 | <https://www.youtube.com/watch?v=WSpNG8aXODE> |

## Process for a year

Each year produces three files in `year_{year}/`, in this order.

### Raw transcript

`recap{year}_transcript_raw.txt`: the auto-generated captions, scripted.
Set `VIDEO_URL` and `OUTPUT` in [youtube_transcript.py](../youtube_transcript.py) and run it.

### Cleaned transcript

`recap{year}_transcript.md`: done by an LLM reading the raw file.

- Join the captions into sentences and paragraphs, with timestamps removed and wording kept raw.
  Some years' captions have no punctuation at all,
  so the sentences have to be rebuilt while reading.
- One `###` section per set release, with bullets for
  release date, set type, major strategies and impact.
- One `####` subsection per event (YCS, WCQ, Worlds), ban list or side topic,
  plus `### Intro` and `### Outro`.
- Correct every card name against the card database,
  following [card_name_lookup.md](../card_name_lookup.md).
  Context matters: the same garble can mean different cards in different sentences,
  so this is judgment, not a find-and-replace list.
- The preamble lists the non-obvious fixes (garbled to canonical)
  and the names left unresolved.

### Archetypes summary

`recap{year}_archetypes.md`: a summary of the cleaned transcript.

- A summary table (with a column for the World Championship result and a line naming the winner),
  then one `##` section per top archetype
  (so the archetypes show in the document outline),
  ordered by how long it was a top deck, longest first.
- Each archetype has its top period (months), new archetype cards or new strong support by set,
  key non-archetype cards (synergy or the reason it topped), results, and what ended it.
- `## Honorable mentions`, one `###` per archetype:
  strong new cards and playable, but never a top deck that year.
- Name archetypes by their real archetype names;
  a community nickname ("Pepe", "PK Fire") appears at most once, in parentheses.
- Every backticked name must match `cards.card_name_en` exactly;
  check them all against the database before finishing.

## Notes and pitfalls

Lessons from earlier years, so they don't need to be repeated as corrections.

- Raw captions go to a `.txt` file, not `.md`:
  thousands of caption lines render as one paragraph in Markdown.
- `youtube_transcript.py` takes its inputs from the constants at the top of the file,
  not from command line arguments.
- A card the video never mentions belongs to another year's notes,
  even if it feels like part of the story
  (`That Grass Looks Greener` pushed Paleozoic in 2017, not 2016).
- Canonical names often differ from what the speaker says or what sounds right:
  `Quickdraw Synchron` (no hyphen), `Kozmoll Dark Lady`, `Gear Gigant X`, `Bujintei Tsukuyomi`,
  `Gold Gadget` and `Silver Gadget` (two cards, not "Golden Silver Gadget").
  Look every name up instead of trusting memory.
- Captions misread common words as card names too
  ("Blue-Eyes" was really "Blue Layer" next to Emergency Teleport),
  so a name that exists can still be the wrong card.
- Event dates are mostly relative ("two weeks later"),
  so top periods are approximate months; say so in the summary.
- The archetypes summary also lists the non-archetype cards that made a deck top,
  not only its archetype cards.
- Writing style for every file here:
  bullets rather than numbered lists, line breaks at semantic boundaries,
  no line over 100 characters outside tables and code blocks.
