# youtube_analysis

Notes on Yu-Gi-Oh! YouTube videos, one directory per creator or series:

- `josh_yugioh_improve/`: Joshua Schmidt's series on getting better at the game.
- `thelawygo_tcg_history_recap/`: TheLawYGO's yearly TCG recaps,
  made by following [yearly_recap_process_guide.md](thelawygo_tcg_history_recap/yearly_recap_process_guide.md).

Card names in auto-generated captions are often garbled;
[card_name_lookup.md](card_name_lookup.md) explains how to correct them against the card database.

## Get the raw transcript

`youtube_transcript.py` saves the raw transcript (captions) of a video,
the starting point for a note:

```bash
pip install -r youtube_analysis/requirements.txt
python youtube_analysis/youtube_transcript.py
# Output file:
# Source: https://www.youtube.com/watch?v=y8S7Ns89Q8A
#
# 0:00:04 When we last left off, the meta was in a
# 0:00:06 state of flux following the release of
# ...
```

Set the inputs at the top of the script (no command line arguments):

- `VIDEO_URL`: a video URL (watch, youtu.be, shorts, embed, live) or a bare video id.
- `OUTPUT`: the file to write, overwritten if it exists.
  A relative path is relative to `youtube_analysis/`,
  for example `thelawygo_tcg_history_recap/year_2016/recap2016_transcript_raw.txt`.
- `LANGUAGES`: language codes in order of preference, default `["en"]`.

Output details:

- The first line is `Source: <VIDEO_URL>`,
  then each caption segment is one line, prefixed with its start time.
- A manually written track is preferred over an auto-generated one in the same language.
  Which track was used is logged to the console.
- If the video has no transcript in `LANGUAGES`,
  the script exits with status 1 and lists the languages that do exist.

It uses [youtube-transcript-api](https://github.com/jdepoix/youtube-transcript-api),
which reads the same captions as the "Show transcript" panel on YouTube.
No API key is needed.
