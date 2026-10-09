# TheLawYGO TCG History Recap

Notes on [TheLawYGO](https://www.youtube.com/@TheLawYGO)'s yearly
"The {year} Yugioh TCG Recap" videos, one directory per year.

To find a year's video, start from the channel's
[Yugioh Yearly Recaps](https://www.youtube.com/playlist?list=PLs8wjLMVLA5T8dgIQfn1TFKZAmJ0gx6oy)
playlist, then search YouTube more broadly
(2018, `WNBlM3AnXhE`, is on the channel but not in the playlist).
If the video is still not found, stop: there is no recap for that year
(none for 2019 to 2021).

| Year | Video                                         |
|------|-----------------------------------------------|
| 2015 | <https://www.youtube.com/watch?v=FLH0yB32WOc> |
| 2016 | <https://www.youtube.com/watch?v=y8S7Ns89Q8A> |
| 2017 | <https://www.youtube.com/watch?v=44Otn2Uo0_8> |
| 2022 | <https://www.youtube.com/watch?v=WSpNG8aXODE> |
| 2023 | <https://www.youtube.com/watch?v=cdcT6BlbbLQ> |
| 2024 | <https://www.youtube.com/watch?v=ns0SiTgcsos> |
| 2025 | <https://www.youtube.com/watch?v=o_ZY7ZvKAu0> |

## Process for a year

Each year produces three files in `year_{year}/`, in this order:
each file is written from the previous one, so finish one before starting the next.

### Raw transcript

`recap{year}_transcript_raw.txt`: the auto-generated captions, scripted.
Set `VIDEO_URL` and `OUTPUT` in [youtube_transcript.py](../youtube_transcript.py) and run it.
Setup and options are in [youtube_transcript_guide.md](../youtube_transcript_guide.md).
When several years run in parallel, each runs its own copy of the script
with an absolute `OUTPUT` path, so the shared constants are not overwritten.

### Cleaned transcript

`recap{year}_transcript.md`: written by the agent running this process,
reading the raw file itself, not by a script, an external model API or a child agent.
Card names need judgment from the whole video's context,
which a split or delegated pass would lose.

File skeleton:

```
# The {year} Yugioh TCG Recap

Author: TheLawYGO.
Source: <{video URL}>
Raw captions: [recap{year}_transcript_raw.txt](recap{year}_transcript_raw.txt)

## Transcript

{preamble: how the captions were cleaned, the card name fixes, "Not resolved, kept as spoken:"}

### Intro

### {set name}

- Release date: {date}
- Set type: {core set, deck building set, ...}
- Major strategies: {archetypes}
- Impact: {the speaker's one-line impact}

{the speaker's text about the set}

#### {event, ban list or side topic}

### Outro
```

- Join the captions into sentences and paragraphs, with timestamps removed and wording kept raw.
  Older captions (2017) have no punctuation at all, so the sentences are rebuilt while reading;
  newer captions are punctuated but garble more names, so the work is mostly name repair.
- Drop caption tags such as "[Music]" and "foreign" without listing them.
  Keep the speaker's shorthand ("OG", "Imperm") and nicknames as spoken.
- An event goes under the `###` of the latest set released before it,
  including an event on the same weekend as a release.
  A reprint-only set still gets its `###` and bullets, with "Major strategies: none".
  For a set released on different dates per region, use the TCG English date
  and mention the other dates in the text.
- Set names and dates come from the card database, in headings and in the summary,
  even when the speaker garbles them (the speaker's text keeps the spoken words).
  - Name: `sets.name_en`, written in title case ("Invasion: Vengeance", not "In Vengeance").
  - Date: the earliest `set_cards.release_date` among the set's `-EN` card numbers;
    `sets.release_date` is the OCG date for a set printed in both games.
  - For a core booster, that date is the sneak peek,
    and the video usually gives the official release about a week later: either is fine.
    When the video is further off, use the database date and add "(the video says ...)".
  - Events keep the video's order either way.
- Correct every card name against the card database,
  following [card_name_lookup.md](../card_name_lookup.md).
  Context matters: the same garble can mean different cards in different sentences,
  so this is judgment, not a find-and-replace list.
  Every confirmed full card name in the text gets backticks;
  the speaker's shorthand ("Imperm") stays plain.
- The preamble lists the non-obvious fixes (garbled to canonical),
  the names inferred only from effect text (marked as inferred),
  and the names left unresolved.

### Archetypes summary

`recap{year}_archetypes.md`: a summary of the cleaned transcript, by the same agent.

File skeleton:

```
# {year} TCG: Top Archetypes

Summary of [recap{year}_transcript.md](recap{year}_transcript.md)
(TheLawYGO, The {year} Yugioh TCG Recap).

{how the archetypes are ordered, and that months are approximate}

| Archetype | Top period | Months | Highlights |

World Championship {year} winner: {deck}, {player} ({country}).

{year} sets in release order, for reference:
{set (date), set (date), ...}

## {top archetype}

- Top period: ...
- New archetype or new strong support: ...
- Key non-archetype cards: ...
- Results: ...
- Ended by: ...

## Honorable mentions

### {archetype}
```

- Months are written in three letters everywhere in the summary ("Jan", "Sep").
- Months counts both ends: Feb to Sep is 8 months.
  Event dates are mostly relative ("two weeks later"), so months are approximate.
- Highlights holds what the month count hides, left empty when there is nothing:
  World Champion (with the player), any World Championship placement the video gives
  (second place, top 4, top 8), "Not legal at Worlds (TCG exclusive at the time)"
  when the video says a top deck was locked out (exclusives reach the other game later),
  dominance such as "Tier zero (Nov to Dec)",
  and wider impact such as "Combined with Ryzeal, then with Yummy".
  Write "tier zero" or "best deck" only when the video says it.
- The `##` sections follow the table order:
  longest top period first, then the earlier start month first when the months are equal.
- A top archetype won a major event (YCS, WCQ, Worlds)
  or made the top cut (whatever its size, top 32 or top 64) at two or more of them,
  even within one month;
  a World Champion deck is always one, even with no TCG results.
  The top period counts the months the deck was topping:
  a stretch where it dropped out of the meta (a ban list knocked it out for months)
  is left out, as in "Jan to Feb, Sep to Oct" (4 months),
  while a dip of about a month still counts.
  Topping is meant loosely: a month counts once the video shows the deck
  impactful enough that the competitive scene treats it as a threat,
  such as "tier 2 to 3 success" at a named event or "already the top deck".
  Anything else goes under `## Honorable mentions`.
- Key non-archetype cards are the synergy or the reason the deck topped.
- A deck built around a non-archetype engine (Mystic Mine) can be a top deck;
  name it by the engine.
- Name archetypes by their real archetype names, with the canonical spelling
  ("Snake-Eye", not "Snake Eyes");
  a community nickname ("Pepe", "PK Fire") appears at most once, in parentheses.
  This applies to the summary; the cleaned transcript keeps the speaker's wording.

## Notes and pitfalls

Lessons from earlier years, so they don't need to be repeated as corrections.

- Backticks are for confirmed card names only, and every one must match `cards.card_name_en`.
  Archetype, engine, set, event and player names stay plain;
  a guessed name stays in quotes as spoken.
- Player and event names in captions are garbled too ("Ys Squad laara" is YCS Guadalajara).
  Fix event names from context.
  Only the World Champion's name is checked against an external source
  (noting "heard as ..." when it differs);
  other player names stay as heard, and the preamble says so.
- A card the video never mentions belongs to another year's notes,
  even if it feels like part of the story
  (`That Grass Looks Greener` pushed Paleozoic in 2017, not 2016).
- Canonical names often differ from what the speaker says or what sounds right:
  `Quickdraw Synchron` (no hyphen), `Kozmoll Dark Lady`, `Gear Gigant X`, `Bujintei Tsukuyomi`,
  `Gold Gadget` and `Silver Gadget` (two cards, not "Golden Silver Gadget").
  Newer names add more traps:
  symbols (`Live☆Twin Lil-la`, `Backup @Ignister`, `Maliss <P> Dormouse`),
  singular versus plural (`Exosister Martha` but `Exosisters Magnifica`),
  a leading "The" (`The Arrival Cyberse @Ignister`),
  and one garble for two cards ("rise heart": `Kashtira Riseheart` or `Kashtira Arise-Heart`).
  Look every name up instead of trusting memory.
- Captions misread common words as card names too:
  "Blue-Eyes" was really `Super Quantum Blue Layer`,
  given away by the rank three and Super Quant context,
  so a name that exists can still be the wrong card.
- The World Championship is usually in August (except 2024: September 7th to 8th, Seattle),
  and it uses a combined TCG and OCG ban list,
  so the winning deck can differ from the TCG decks topping at the same time.
  Use the date to place Worlds among the video's events,
  then fact check the winner, the deck and the player name against an external source
  (Yugipedia, roadoftheking.com): player names in captions are often garbled.
  Yugipedia blocks plain page fetches; its API works with curl:
  `https://yugipedia.com/api.php?action=parse&page={page}&prop=wikitext&format=json`.
  There was no World Championship from 2020 to 2022 (COVID-19): write "Not held".
