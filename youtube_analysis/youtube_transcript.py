"""Saves the raw transcript (captions) of a YouTube video to a file.

Usage: set VIDEO_URL and OUTPUT (and LANGUAGES if needed) below, then run
    python youtube_transcript.py

OUTPUT starts with a "Source: <VIDEO_URL>" line,
then one caption segment per line as "mm:ss text"
("h:mm:ss text" for videos of an hour or more).
A manually written track is preferred over
an auto-generated one when both exist in the same language.
Exits with status 1 if the video has no transcript in LANGUAGES.
"""
import os
import re
import sys
from urllib.parse import parse_qs, urlparse

from youtube_transcript_api import CouldNotRetrieveTranscript, YouTubeTranscriptApi


HERE = os.path.dirname(os.path.abspath(__file__))

# Inputs (edit for each run):

# A watch, youtu.be, shorts, embed or live URL, or a bare video id.
VIDEO_URL = "https://www.youtube.com/watch?v=WSpNG8aXODE"
# Overwritten if it exists. A relative path is relative to this script's directory.
OUTPUT = "thelawygo_tcg_history_recap/year_2022/recap2022_transcript_raw.txt"
# Language codes in order of preference.
LANGUAGES = ["en"]


def main():
    video_id = parse_video_id(VIDEO_URL)

    api = YouTubeTranscriptApi()
    try:
        transcript_list = api.list(video_id)
        transcript = transcript_list.find_transcript(LANGUAGES)
        segments = transcript.fetch()
    except CouldNotRetrieveTranscript as err:
        print(f"error fetching transcript for {video_id}: {err}", file=sys.stderr)
        sys.exit(1)

    kind = "auto-generated" if transcript.is_generated else "manual"
    print(f"using {kind} transcript, language {transcript.language_code}, "
          f"{len(segments)} segments", file=sys.stderr)

    is_long_video = len(segments) > 0 and segments[-1].start >= 3600
    lines = [f"Source: {VIDEO_URL}", ""]
    for segment in segments:
        text = " ".join(segment.text.split())
        lines.append(f"{format_timestamp(segment.start, is_long_video)} {text}")

    output_path = os.path.normpath(os.path.join(HERE, OUTPUT))
    os.makedirs(os.path.dirname(output_path), exist_ok=True)
    with open(output_path, "w", encoding="utf-8", newline="\n") as f:
        f.write("\n".join(lines) + "\n")
    print(f"wrote {output_path}", file=sys.stderr)


def parse_video_id(url_or_id):
    """Accepts a watch, youtu.be, shorts, embed or live URL, or a bare video id."""
    if VIDEO_ID_PATTERN.match(url_or_id):
        return url_or_id
    parsed = urlparse(url_or_id)
    query_ids = parse_qs(parsed.query).get("v")
    if query_ids:
        return query_ids[0]
    path_parts = [part for part in parsed.path.split("/") if part]
    if path_parts and VIDEO_ID_PATTERN.match(path_parts[-1]):
        return path_parts[-1]
    print(f"error: cannot find a video id in {url_or_id!r}", file=sys.stderr)
    sys.exit(2)


def format_timestamp(seconds, is_long_video):
    total = int(seconds)
    hours, remainder = divmod(total, 3600)
    minutes, secs = divmod(remainder, 60)
    if is_long_video:
        return f"{hours}:{minutes:02d}:{secs:02d}"
    return f"{minutes:02d}:{secs:02d}"


# YouTube video ids are always 11 characters from this set.
VIDEO_ID_PATTERN = re.compile(r"^[A-Za-z0-9_-]{11}$")

if __name__ == "__main__":
    main()
