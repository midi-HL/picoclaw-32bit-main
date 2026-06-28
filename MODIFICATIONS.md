[中文版](docs/project/MODIFICATIONS.zh.md) | [日本語](docs/project/MODIFICATIONS.ja.md) | [한국어](docs/project/MODIFICATIONS.ko.md) | [Português](docs/project/MODIFICATIONS.pt-br.md) | [Tiếng Việt](docs/project/MODIFICATIONS.vi.md) | [Français](docs/project/MODIFICATIONS.fr.md) | [Italiano](docs/project/MODIFICATIONS.it.md) | [Bahasa Indonesia](docs/project/MODIFICATIONS.id.md) | [Malay](docs/project/MODIFICATIONS.ms.md) | **English**

# Modifications from Upstream PicoClaw

This document describes all modifications made in this fork (`picoclaw-32bit-main`) compared to the upstream [PicoClaw](https://github.com/sipeed/picoclaw) project.

---

## Table of Contents

- [1. MiMo Multimodal Support](#1-mimo-multimodal-support)
- [2. Video Understanding](#2-video-understanding)
- [3. Telegram Video Messages](#3-telegram-video-messages)
- [4. load_video Tool](#4-load_video-tool)
- [5. Audio Data URL Encoding](#5-audio-data-url-encoding)
- [6. Config Robustness](#6-config-robustness)
- [7. API Changes](#7-api-changes)
- [8. 32-bit Platform Support](#8-32-bit-platform-support)
- [Known Limitations](#known-limitations)

---

## 1. MiMo Multimodal Support

**Problem:** The upstream provider code sends audio in OpenAI format (`{"data": "base64", "format": "wav"}`), but MiMo's API expects a full data URL (`{"data": "data:audio/wav;base64,..."}`).

**Changes:**
- `pkg/providers/common/common.go` — Added `SerializeOptions` struct and `SerializeMessagesWithOptions` function with provider-aware format selection:
  - **Audio**: MiMo → full data URL in `data` field; Standard → split base64 + `format` field
  - **Video**: MiMo → `video_url` with `fps`/`media_resolution`; Standard → skipped (no standard type)
  - **Image**: Always standard `image_url` (universal)
- `pkg/providers/openai_compat/provider.go` — Now passes `providerName` and `apiBase` to `SerializeMessagesWithOptions`.

**Format compatibility matrix:**

| Type | MiMo Format | Standard OpenAI Format | Works With |
|------|------------|----------------------|------------|
| Image | `image_url` + data URL | `image_url` + data URL | All multimodal models |
| Audio | `input_audio.data` = full data URL | `input_audio.data` = base64 + `format` field | MiMo / OpenAI |
| Video | `video_url` + `fps` + `media_resolution` | No standard type | MiMo only |

**Backward compatible:** `SerializeMessages()` still works with standard OpenAI format for callers that don't pass options.

---

## 2. Video Understanding

**Problem:** The `video_model` configuration field existed but was never used by the agent code.

**Changes:**
- `pkg/agent/instance.go` — Added `VideoCandidates` field to `AgentInstance` and resolved `video_model` candidates at startup.
- `pkg/agent/llm_media.go` — Added `describeVideoProxy()` function implementing a **delegation pattern**:
  1. Detects `data:video/` URLs in the current turn
  2. Sends video + description prompt to `video_model`
  3. Injects the description into the message content as `[系统消息：以下是用户发送视频的描述]`
  4. Main model receives the description and replies normally
- `pkg/agent/llm_media.go` — `routeMediaTurn` calls `describeVideoProxy` before falling back to image model routing.

**Flow:**
```
User sends video
  → video_model describes video
  → Description injected into message
  → Main model replies using description as context
```

---

## 3. Telegram Video Messages

**Problem:** `collectTelegramMessageParts` handled Photo, Voice, Audio, and Document — but not Video. Video messages were silently dropped.

**Changes:**
- `pkg/channels/telegram/telegram.go` — Added `msg.Video` handling: downloads video file, stores in media store, adds `[video]` tag to message content.

---

## 4. load_video Tool

**New feature:** A tool that allows the AI to load and analyze local video files.

**Files:**
- `pkg/tools/fs/load_video.go` — New tool implementation (mirrors `load_image`).
- `pkg/tools/fs_facade.go` — Added `LoadVideoTool` type alias and `NewLoadVideoTool` factory.
- `pkg/agent/agent_init.go` — Registered `load_video` tool.
- `pkg/config/config.go` — Added `LoadVideo ToolConfig` field.
- `pkg/agent/context.go` — Updated multimodal system prompt to mention `load_video`.

**How it works:**
1. AI calls `load_video(path="video.mp4")`
2. Tool validates path, detects MIME type, stores in media store
3. Returns `media://` reference
4. `resolveMediaRefs` encodes as `data:video/mp4;base64,...`
5. Provider sends as `video_url` format to MiMo

---

## 5. Audio Data URL Encoding

**Problem:** Audio from user messages was not encoded as data URLs for the model.

**Changes:**
- `pkg/agent/agent_media.go` — `resolveMediaRefs` now encodes audio and video as data URLs for both user messages and tool results.
- `pkg/agent/prompt_turn.go` — `toolImageFollowUpPromptMessage` detects video data URLs and updates the synthetic user message text accordingly.

---

## 6. Config Robustness

### Unknown fields are warnings, not errors

**Problem:** Config files with deprecated fields (e.g., from older versions) caused startup failures.

**Changes:**
- `pkg/config/diagnostics.go` — `decodeJSONWithDiagnostics` now logs unknown fields as warnings to stderr instead of returning errors.

### Config API request body limit

**Problem:** The PATCH/PUT `/api/config` endpoint had a 1MB body limit, too small for voice clone audio data in base64.

**Changes:**
- `web/backend/api/config.go` — Increased body limit from 1MB to 20MB for both PUT and PATCH handlers.

### VoiceConfig MimoConfig field

**Problem:** The Go `VoiceConfig` struct had no `MimoConfig` field, so MiMo-specific settings were lost during JSON round-trip.

**Changes:**
- `pkg/config/config.go` — Added `VoiceMimoConfig` struct with ASR fields (`asr_provider`, `asr_language`, `asr_api_key`).

---

## 7. API Changes

This fork adds the following API capabilities. For detailed API documentation, see [API Reference](docs/api/README.md).

### Chat API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/chat` | POST | Synchronous chat — send message, receive complete reply |
| `/api/chat/stream` | POST | Streaming chat — SSE real-time token output |

### New Tool: load_video

The `load_video` tool is registered as a callable tool for the AI agent. It accepts a file path and returns a video analysis reference.

**Tool parameters:**
```json
{
  "path": "path/to/video.mp4"
}
```

**Tool result:**
```
Video loaded: video.mp4
[video: /path/to/video.mp4]
```

The video is automatically encoded as a data URL and sent to the model in `video_url` format.

---

## 8. 32-bit Platform Support

> This section documents the 32-bit platform support that was added to this fork.

### Supported 32-bit targets

| OS | GOARCH | Binary Name | Minimum OS Version |
|----|--------|-------------|-------------------|
| Linux | `386` | `picoclaw-linux-386` | Any 32-bit Linux with kernel 2.6+ |
| Linux | `arm` (GOARM=7) | `picoclaw-linux-arm` | ARMv7 Linux (e.g. Raspberry Pi) |
| Linux | `mipsle` | `picoclaw-linux-mipsle` | MIPS32 little-endian Linux (softfloat) |
| Linux | `mips` | `picoclaw-linux-mips` | MIPS32 big-endian Linux (softfloat) |
| Windows | `386` | `picoclaw-windows-386.exe` | Windows XP SP3 / Vista / 7 / 8 / 8.1 / 10 (32-bit) |

### What was changed

- Added `linux/386`, `linux/arm`, `linux/mipsle`, `linux/mips` build targets to the Makefile
- The `windows/386` target was already present in the Makefile and `.goreleaser.yaml`
- Source files using modernc sqlite/libc have build tags excluding both `mipsle` and `mips` big-endian

### Implementation details

- Uses the pure-Go olm cryptographic implementation via the `goolm` build tag (no CGO / `libolm` dependency)
- All Windows APIs used are Vista/Win7 level — no Win10+ exclusive APIs
- `unsafe.Pointer` usage is architecture-neutral
- Feishu/Lark channel is **not available** on 32-bit (upstream SDK limitation, handled gracefully at runtime)
- Matrix channel is **not available** on MIPS (both LE and BE) due to modernc sqlite/libc build constraints
- MIPS targets use `GOMIPS=softfloat` for compatibility with FP-lacking kernels

### Build from source

```bash
# Linux 32-bit x86
CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -v -tags goolm,stdjson -o build/picoclaw-linux-386 ./cmd/picoclaw

# Linux 32-bit ARM (GOARM=7)
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -v -tags goolm,stdjson -o build/picoclaw-linux-arm ./cmd/picoclaw

# Linux 32-bit MIPS LE (softfloat, no goolm)
CGO_ENABLED=0 GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -v -tags stdjson -o build/picoclaw-linux-mipsle ./cmd/picoclaw

# Linux 32-bit MIPS BE (softfloat, no goolm)
CGO_ENABLED=0 GOOS=linux GOARCH=mips GOMIPS=softfloat go build -v -tags stdjson -o build/picoclaw-linux-mips ./cmd/picoclaw

# Windows 32-bit (cross-compile from any OS)
CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -v -tags goolm,stdjson -o build/picoclaw-windows-386.exe ./cmd/picoclaw

# Or use the Makefile targets (builds all platforms including 32-bit):
make build-all
```

### Runtime system requirements

| Resource | Minimum |
|----------|---------|
| CPU | Any x86 processor with SSE2 support |
| RAM | 512 MB |
| Disk | 100 MB (binary) + workspace storage |
| Network | Internet access for LLM API calls |

---

## Known Limitations

### Video Format is MiMo-Only

The `video_url` format used for video input is MiMo-specific. There is no standard OpenAI video type. When using non-MiMo models, video content is silently skipped.

**Workaround:** Configure `agents.defaults.video_model` to a MiMo model for video analysis — the delegation pattern will describe the video and pass the text to the main model.

### Chat API Does Not Accept Multimodal Input

The `/api/chat` endpoint accepts only plain text messages (`{"message": "text"}`). It does not support the OpenAI Messages API format with multipart content (images, audio, video inline). Multimodal content is only supported when sent through channel integrations (Telegram, Discord, etc.) or internal tool results.
