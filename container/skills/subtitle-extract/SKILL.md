---
name: subtitle-extract
description: Extract subtitles and transcripts from YouTube and Instagram video URLs. Use when the user provides a video URL and wants the transcript, subtitles, or spoken content extracted. Supports YouTube (auto-captions, manual subs) and Instagram Reels. Falls back to Whisper API transcription when no subtitles are available.
allowed-tools: Bash(yt-dlp:*), Bash(curl:*), Bash(cat:*), Bash(ffmpeg:*)
---

# Subtitle & Transcript Extraction

Extract subtitles from video URLs using yt-dlp. Falls back to Whisper API when no subtitles exist.

## Setup

Source the API key before using Whisper:
```bash
source /workspace/group/.env
```

## Quick Reference

```bash
# YouTube — extract existing subtitles (auto-generated or manual)
yt-dlp --write-auto-sub --write-sub --sub-lang ko,en --skip-download --sub-format vtt -o "/tmp/subs/%(title)s" "URL"
cat /tmp/subs/*.vtt

# Instagram/YouTube — download audio + transcribe with Whisper API
source /workspace/group/.env
yt-dlp -x --audio-format mp3 --audio-quality 5 -o "/tmp/audio/%(title)s.%(ext)s" "URL"
curl -s https://api.openai.com/v1/audio/transcriptions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -F file=@/tmp/audio/*.mp3 \
  -F model=whisper-1 \
  -F language=ko \
  -F response_format=text
```

## Strategy

1. **Try subtitles first** (free, instant):
   ```bash
   mkdir -p /tmp/subs
   yt-dlp --write-auto-sub --write-sub --sub-lang ko,en,ja --skip-download --sub-format vtt -o "/tmp/subs/%(id)s" "URL" 2>&1
   ```
   If `.vtt` files are created → parse and return the text.

2. **Fall back to Whisper API** (when no subs available):
   ```bash
   mkdir -p /tmp/audio
   yt-dlp -x --audio-format mp3 --audio-quality 5 -o "/tmp/audio/%(id)s.%(ext)s" "URL"

   AUDIO_FILE=$(ls /tmp/audio/*.mp3 | head -1)
   curl -s https://api.openai.com/v1/audio/transcriptions \
     -H "Authorization: Bearer $OPENAI_API_KEY" \
     -F file=@"$AUDIO_FILE" \
     -F model=whisper-1 \
     -F language=ko \
     -F response_format=text
   ```

3. **Clean up** after extraction:
   ```bash
   rm -rf /tmp/subs /tmp/audio
   ```

## Parsing VTT Subtitles

VTT files contain timestamps + text. To extract plain text:

```bash
# Remove timestamps and metadata, keep only text lines
cat /tmp/subs/*.vtt | grep -v "^$" | grep -v "^WEBVTT" | grep -v "^Kind:" | grep -v "^Language:" | grep -v "^[0-9]" | grep -v "^\-\->" | sed 's/<[^>]*>//g' | awk '!seen[$0]++'
```

## Supported Platforms

| Platform | Subtitles | Audio Download | Notes |
|----------|-----------|---------------|-------|
| YouTube | Auto-captions + manual subs | Yes | Best subtitle support |
| Instagram Reels | Rarely available | Yes | Usually needs Whisper fallback |
| TikTok | Sometimes | Yes | May need cookies for some videos |

## Tips

- Instagram URLs: use the full URL (e.g., `https://www.instagram.com/reel/ABC123/`)
- YouTube: both `youtu.be` and `youtube.com` URLs work
- For long videos (>25MB audio), split before sending to Whisper API:
  ```bash
  ffmpeg -i input.mp3 -f segment -segment_time 600 -c copy /tmp/audio/chunk_%03d.mp3
  ```
- Whisper API limit: 25MB per file, ~$0.006/minute of audio
