[English](../../MODIFICATIONS.md) | [日本語](MODIFICATIONS.ja.md) | [한국어](MODIFICATIONS.ko.md) | [Português](MODIFICATIONS.pt-br.md) | [Tiếng Việt](MODIFICATIONS.vi.md) | [Français](MODIFICATIONS.fr.md) | [Italiano](MODIFICATIONS.it.md) | [Bahasa Indonesia](MODIFICATIONS.id.md) | [Malay](MODIFICATIONS.ms.md) | **中文**

# 相对于上游 PicoClaw 的修改说明

本文档描述此 fork（`picoclaw-32bit-main`）相对于上游 [PicoClaw](https://github.com/sipeed/picoclaw) 项目的所有修改。

---

## 目录

- [1. MiMo 多模态支持](#1-mimo-多模态支持)
- [2. 视频理解](#2-视频理解)
- [3. Telegram 视频消息](#3-telegram-视频消息)
- [4. load_video 工具](#4-load_video-工具)
- [5. 音频 Data URL 编码](#5-音频-data-url-编码)
- [6. 配置健壮性](#6-配置健壮性)
- [7. API 变更](#7-api-变更)
- [已知限制](#已知限制)
- [8. 32 位平台支持](#8-32-位平台支持)

---

## 1. MiMo 多模态支持

**问题：** 上游 provider 代码使用 OpenAI 格式发送音频（`{"data": "base64", "format": "wav"}`），但 MiMo API 要求完整 data URL（`{"data": "data:audio/wav;base64,..."}`）。

**修改内容：**
- `pkg/providers/common/common.go` — `SerializeMessages` 现在在 `input_audio.data` 字段中发送完整 `data:` URL，而不是拆分为 `data` + `format` 字段。
- `pkg/providers/common/common.go` — 新增 `video_url` 格式支持 MiMo 视频输入：`{"type": "video_url", "video_url": {"url": "data:video/mp4;base64,..."}, "fps": 2}`。

**MiMo API 格式参考：**

| 类型 | 格式 |
|------|------|
| 音频 | `{"type": "input_audio", "input_audio": {"data": "data:audio/wav;base64,..."}}` |
| 视频 | `{"type": "video_url", "video_url": {"url": "data:video/mp4;base64,..."}, "fps": 2, "media_resolution": "default"}` |
| 图片 | `{"type": "image_url", "image_url": {"url": "data:image/png;base64,..."}}` |

---

## 2. 视频理解

**问题：** `video_model` 配置字段存在但从未被 agent 代码使用。

**修改内容：**
- `pkg/agent/instance.go` — `AgentInstance` 新增 `VideoCandidates` 字段，启动时解析 `video_model` 候选。
- `pkg/agent/llm_media.go` — 新增 `describeVideoProxy()` 函数，实现**代理转述模式**：
  1. 检测当前 turn 中的 `data:video/` URL
  2. 将视频 + 描述提示词发送给 `video_model`
  3. 将描述注入消息内容：`[系统消息：以下是用户发送视频的描述]`
  4. 主模型基于描述回复用户
- `pkg/agent/llm_media.go` — `routeMediaTurn` 在回退到 image model 路由之前先调用 `describeVideoProxy`。

**流程：**
```
用户发送视频
  → video_model 描述视频
  → 描述注入消息
  → 主模型基于描述回复
```

---

## 3. Telegram 视频消息

**问题：** `collectTelegramMessageParts` 处理了 Photo、Voice、Audio 和 Document，但没有处理 Video。视频消息被静默丢弃。

**修改内容：**
- `pkg/channels/telegram/telegram.go` — 新增 `msg.Video` 处理：下载视频文件、存入 media store、在消息内容中添加 `[video]` 标签。

---

## 4. load_video 工具

**新功能：** 允许 AI 加载和分析本地视频文件的工具。

**相关文件：**
- `pkg/tools/fs/load_video.go` — 新工具实现（与 `load_image` 同构）。
- `pkg/tools/fs_facade.go` — 新增 `LoadVideoTool` 类型别名和 `NewLoadVideoTool` 工厂函数。
- `pkg/agent/agent_init.go` — 注册 `load_video` 工具。
- `pkg/config/config.go` — 新增 `LoadVideo ToolConfig` 字段。
- `pkg/agent/context.go` — 更新多模态系统提示词，提及 `load_video`。

**工作原理：**
1. AI 调用 `load_video(path="video.mp4")`
2. 工具验证路径、检测 MIME 类型、存入 media store
3. 返回 `media://` 引用
4. `resolveMediaRefs` 编码为 `data:video/mp4;base64,...`
5. Provider 以 `video_url` 格式发送给 MiMo

---

## 5. 音频 Data URL 编码

**问题：** 用户消息中的音频未被编码为 data URL。

**修改内容：**
- `pkg/agent/agent_media.go` — `resolveMediaRefs` 现在为用户消息和工具结果中的音频/视频编码为 data URL。
- `pkg/agent/prompt_turn.go` — `toolImageFollowUpPromptMessage` 检测视频 data URL 并相应更新合成用户消息文本。

---

## 6. 配置健壮性

### 未知字段降级为警告

**问题：** 包含已弃用字段的配置文件（如旧版本）导致启动失败。

**修改内容：**
- `pkg/config/diagnostics.go` — `decodeJSONWithDiagnostics` 现在将未知字段记录为 stderr 警告，而不是返回错误。

### 配置 API 请求体限制

**问题：** PATCH/PUT `/api/config` 端点的请求体限制为 1MB，对于 base64 编码的音色复刻音频数据来说太小。

**修改内容：**
- `web/backend/api/config.go` — PUT 和 PATCH 处理器的请求体限制从 1MB 提高到 20MB。

### VoiceConfig MimoConfig 字段

**问题：** Go 的 `VoiceConfig` 结构体没有 `MimoConfig` 字段，导致 MiMo 特定设置在 JSON 序列化/反序列化时丢失。

**修改内容：**
- `pkg/config/config.go` — 新增 `VoiceMimoConfig` 结构体，包含 ASR 字段（`asr_provider`、`asr_language`、`asr_api_key`）。

---

## 7. API 变更

此 fork 新增了以下 API 能力。详细 API 文档请参见 [API 参考文档](../api/README.zh.md)。

### 聊天 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/chat` | POST | 同步聊天 — 发送消息，接收完整回复 |
| `/api/chat/stream` | POST | 流式聊天 — SSE 实时 token 输出 |

### 新工具：load_video

`load_video` 工具注册为 AI agent 的可调用工具。接受文件路径参数，返回视频分析引用。

**工具参数：**
```json
{
  "path": "path/to/video.mp4"
}
```

**工具结果：**
```
Video loaded: video.mp4
[video: /path/to/video.mp4]
```

视频会自动编码为 data URL，并以 `video_url` 格式发送给模型。

---

## 8. 32 位平台支持

> 本文档记录此 fork 新增的 32 位平台支持。

### 支持的 32 位目标平台

| 操作系统 | GOARCH | 二进制文件名 | 最低系统版本 |
|---------|--------|-------------|------------|
| Linux | `386` | `picoclaw-linux-386` | 任何内核 2.6+ 的 32 位 Linux |
| Linux | `arm` (GOARM=7) | `picoclaw-linux-arm` | ARMv7 Linux（如树莓派） |
| Linux | `mipsle` | `picoclaw-linux-mipsle` | MIPS32 小端序 Linux（软浮点） |
| Linux | `mips` | `picoclaw-linux-mips` | MIPS32 大端序 Linux（软浮点） |
| Windows | `386` | `picoclaw-windows-386.exe` | Windows XP SP3 / Vista / 7 / 8 / 8.1 / 10 (32 位) |

### 修改内容

- 在 Makefile 的 `build-all` 和 `build-whatsapp-native` 目标中新增了 `linux/386`、`linux/arm`、`linux/mipsle`、`linux/mips` 构建目标
- `windows/386` 目标已存在于 Makefile 和 `.goreleaser.yaml` 中
- 使用 modernc sqlite/libc 的源文件已添加构建标签排除 `mipsle` 和 `mips` 大端序

### 实现方式

- 通过 `goolm` 构建标签使用纯 Go 实现的 olm 加密库，无需 CGO / `libolm` 依赖
- 所有使用的 Windows API 均为 Vista/Win7 级别，无 Win10+ 专有 API
- `unsafe.Pointer` 的使用与架构无关
- 飞书/Lark 频道在 32 位平台上**不可用**（上游 SDK 限制，运行时会优雅处理）
- Matrix 频道在 MIPS（LE 和 BE）上**不可用**，受 modernc sqlite/libc 构建限制
- MIPS 目标使用 `GOMIPS=softfloat` 以兼容无浮点单元的内核

### 从源码编译

```bash
# Linux 32 位 x86
CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -v -tags goolm,stdjson -o build/picoclaw-linux-386 ./cmd/picoclaw

# Linux 32 位 ARM (GOARM=7)
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -v -tags goolm,stdjson -o build/picoclaw-linux-arm ./cmd/picoclaw

# Linux 32 位 MIPS 小端序（软浮点，无 goolm）
CGO_ENABLED=0 GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -v -tags stdjson -o build/picoclaw-linux-mipsle ./cmd/picoclaw

# Linux 32 位 MIPS 大端序（软浮点，无 goolm）
CGO_ENABLED=0 GOOS=linux GOARCH=mips GOMIPS=softfloat go build -v -tags stdjson -o build/picoclaw-linux-mips ./cmd/picoclaw

# Windows 32 位（可从任意操作系统交叉编译）
CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -v -tags goolm,stdjson -o build/picoclaw-windows-386.exe ./cmd/picoclaw

# 或使用 Makefile 目标（构建所有平台，包括 32 位）：
make build-all
```

### 运行时系统要求

| 资源 | 最低要求 |
|-----|---------|
| CPU | 任何支持 SSE2 的 x86 处理器 |
| 内存 | 512 MB |
| 磁盘 | 100 MB（二进制文件）+ 工作空间存储 |
| 网络 | 需要互联网访问以调用 LLM API |

---

## 已知限制

### 多模态格式兼容性

Provider 层（`pkg/providers/common/common.go`）使用 **MiMo 专用格式** 发送音频和视频。这意味着：

| 类型 | 当前格式 | 标准 OpenAI 格式 | 兼容性 |
|------|---------|-----------------|--------|
| 图片 | `image_url` + data URL | `image_url` + data URL | ✅ 标准 — 所有多模态模型可用 |
| 音频 | `input_audio.data` = 完整 data URL（`data:audio/wav;base64,...`） | `input_audio.data` = 纯 base64 + 单独 `format` 字段 | ⚠️ MiMo 专用 — 标准 OpenAI 模型可能拒绝 |
| 视频 | `video_url` + data URL + `fps` + `media_resolution` | OpenAI API 中无此类型 | ❌ MiMo 专用 — 其他 provider 不支持 |

**影响：** 使用非 MiMo 的多模态模型（如 GPT-4o、Gemini）时，图片可正常工作，但音频和视频可能失败或被忽略，因为 provider 使用的是 MiMo 专用格式而非标准 OpenAI 格式。

**解决方案：** 使用 MiMo 模型进行音频/视频分析，或通过 `agents.defaults.image_model` 为多模态任务配置单独的模型（仅限图片）。

### 聊天 API 不支持多模态输入

`/api/chat` 端点仅接受纯文本消息（`{"message": "文本"}`），不支持 OpenAI Messages API 的多部分内容格式（内联图片、音频、视频）。多模态内容仅支持通过频道集成（Telegram、Discord 等）或内部工具结果发送。
