[中文版](README.zh.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Português](README.pt-br.md) | [Tiếng Việt](README.vi.md) | [Français](README.fr.md) | [Italiano](README.it.md) | [Bahasa Indonesia](README.id.md) | [Malay](README.ms.md) | **English**

# PicoClaw Fork — API Reference

This document describes the API endpoints available in this fork of PicoClaw, with focus on endpoints useful for third-party client development.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Authentication](#authentication)
- [Chat API (Core)](#chat-api-core)
- [Session Management](#session-management)
- [Model Management](#model-management)
- [Configuration](#configuration)
- [Gateway Lifecycle](#gateway-lifecycle)
- [Health](#health)
- [Tool: load_video](#tool-load_video)

---

## Architecture Overview

PicoClaw runs two HTTP servers:

| Server | Default Port | Purpose |
|--------|-------------|---------|
| **Launcher** | 18800 | Dashboard UI, API proxy, config management |
| **Gateway** | 18790 | Core chat API, health checks, channel webhooks |

For third-party client development, you primarily interact with the **Launcher** on port 18800. The Launcher proxies chat requests to the Gateway automatically.

---

## Authentication

### POST /api/auth/login

Authenticate with the dashboard password.

**Request:**
```json
{
  "password": "your-password"
}
```

**Response:**
```json
{
  "ok": true
}
```

Sets a session cookie for subsequent requests.

### GET /api/auth/status

Check authentication status.

**Response:**
```json
{
  "authenticated": true,
  "password_set": true
}
```

### POST /api/auth/setup

Set or change the dashboard password.

**Request:**
```json
{
  "password": "new-password",
  "confirm": "new-password"
}
```

---

## Chat API (Core)

These are the most important endpoints for third-party clients.

### POST /api/chat

Synchronous chat — send a message, wait for the complete reply.

**Request:**
```json
{
  "message": "Hello, what can you do?",
  "channel": "pico",
  "chat_id": "my-client-user"
}
```

**Response:**
```json
{
  "reply": "Hello! I'm PicoClaw, your AI assistant...",
  "chat_id": "my-client-user"
}
```

**cURL example:**
```bash
curl -X POST http://localhost:18800/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What is 2+2?",
    "channel": "pico",
    "chat_id": "test-user"
  }'
```

### POST /api/chat/stream

Streaming chat — receive reply as Server-Sent Events (SSE).

**Request:** Same as `/api/chat`.

**Response:** SSE stream with event type `message`:
```
data: {"content": "Hello"}
data: {"content": "! "}
data: {"content": "I'm "}
data: {"content": "PicoClaw"}
data: {"done": true}
```

**cURL example:**
```bash
curl -X POST http://localhost:18800/api/chat/stream \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Tell me a joke",
    "channel": "pico",
    "chat_id": "test-user"
  }'
```

---

## Session Management

### GET /api/sessions

List all chat sessions.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `offset` | int | 0 | Pagination offset |
| `limit` | int | 50 | Max results per page |

**Response:**
```json
[
  {
    "id": "session-abc123",
    "channel": "pico",
    "chat_id": "my-client-user",
    "created_at": "2026-06-27T10:00:00Z",
    "updated_at": "2026-06-27T12:30:00Z",
    "message_count": 42
  }
]
```

### GET /api/sessions/{id}

Get full message history for a session.

**Response:**
```json
{
  "id": "session-abc123",
  "messages": [
    {
      "role": "user",
      "content": "Hello",
      "timestamp": "2026-06-27T10:00:00Z"
    },
    {
      "role": "assistant",
      "content": "Hi! How can I help?",
      "timestamp": "2026-06-27T10:00:05Z"
    }
  ]
}
```

### DELETE /api/sessions/{id}

Delete a session and its associated files.

**Response:**
```json
{
  "ok": true
}
```

---

## Model Management

### GET /api/models

List all configured models with availability status.

**Response:**
```json
{
  "models": [
    {
      "model_name": "mimo-v2.5",
      "model": "mimo/mimo-v2.5",
      "provider": "mimo",
      "available": true
    }
  ],
  "default_model": "mimo-v2.5"
}
```

### POST /api/models/test-inline

Test connectivity to a model endpoint without saving.

**Request:**
```json
{
  "provider": "mimo",
  "model": "mimo-v2.5",
  "api_base": "https://api.xiaomimimo.com/v1",
  "api_key": "your-api-key"
}
```

**Response:**
```json
{
  "ok": true,
  "latency_ms": 230,
  "status": "connected"
}
```

---

## Configuration

### GET /api/config

Returns the complete system configuration as JSON.

### PATCH /api/config

Partially update configuration using JSON Merge Patch (RFC 7396). Only specified fields are modified.

**Request:**
```json
{
  "agents": {
    "defaults": {
      "model_name": "mimo-v2.5"
    }
  }
}
```

### PUT /api/config

Replace the entire configuration. Use with caution.

### POST /api/config/reset

Reset configuration to factory defaults. Preserves API keys and security credentials.

---

## Gateway Lifecycle

### GET /api/gateway/status

Check gateway runtime status.

**Response:**
```json
{
  "status": "running",
  "pid": 12345,
  "config_model": "mimo-v2.5",
  "restart_needed": false,
  "start_allowed": true
}
```

### POST /api/gateway/start

Start the gateway subprocess.

### POST /api/gateway/stop

Stop the running gateway.

### POST /api/gateway/restart

Restart the gateway (stop + start).

---

## Health

### GET /health

Returns health status. Always returns 200 when server is running.

**Response:**
```json
{
  "status": "ok",
  "uptime": "2h30m",
  "pid": 12345
}
```

### GET /ready

Returns readiness status. Returns 503 if not ready.

### POST /reload

Trigger a config reload. Protected by optional bearer token.

---

## Tool: load_video

A new tool available to the AI agent for loading and analyzing local video files.

### Tool Definition

| Field | Value |
|-------|-------|
| Name | `load_video` |
| Parameters | `path` (string, required) |

### Usage

When the AI calls `load_video(path="video.mp4")`:

1. Validates the file path is within the workspace
2. Detects MIME type (must be `video/*`)
3. Stores the file in the media store
4. Returns a `media://` reference

The video is then automatically:
- Encoded as `data:video/mp4;base64,...` by `resolveMediaRefs`
- Sent to the model as `video_url` format by the provider
- Processed by multimodal models (e.g., MiMo) that support video input

### Supported Video Formats

MP4, WebM, MOV, AVI, MKV, and other common formats recognized by the Go `filetype` library.

### Size Limits

Controlled by `agents.defaults.max_media_size` in config (default: 20MB).

---

## Multi-language API Documentation

| Language | File |
|----------|------|
| English | [README.md](README.md) |
| 中文 | [README.zh.md](README.zh.md) |
| 日本語 | [README.ja.md](README.ja.md) |
| 한국어 | [README.ko.md](README.ko.md) |
| Português | [README.pt-br.md](README.pt-br.md) |
| Tiếng Việt | [README.vi.md](README.vi.md) |
| Français | [README.fr.md](README.fr.md) |
| Italiano | [README.it.md](README.it.md) |
| Bahasa Indonesia | [README.id.md](README.id.md) |
| Malay | [README.ms.md](README.ms.md) |
