[English](README.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Português](README.pt-br.md) | [Tiếng Việt](README.vi.md) | [Français](README.fr.md) | [Italiano](README.it.md) | [Bahasa Indonesia](README.id.md) | [Malay](README.ms.md) | **中文**

# PicoClaw Fork — API 参考文档

本文档描述此 PicoClaw fork 中可用的 API 端点，重点介绍对第三方客户端开发有用的端点。

---

## 目录

- [架构概览](#架构概览)
- [认证](#认证)
- [聊天 API（核心）](#聊天-api核心)
- [会话管理](#会话管理)
- [模型管理](#模型管理)
- [配置管理](#配置管理)
- [网关生命周期](#网关生命周期)
- [健康检查](#健康检查)
- [工具：load_video](#工具load_video)

---

## 架构概览

PicoClaw 运行两个 HTTP 服务器：

| 服务器 | 默认端口 | 用途 |
|--------|---------|------|
| **Launcher** | 18800 | Dashboard UI、API 代理、配置管理 |
| **Gateway** | 18790 | 核心聊天 API、健康检查、频道 webhook |

对于第三方客户端开发，主要与 **Launcher**（端口 18800）交互。Launcher 会自动将聊天请求代理到 Gateway。

---

## 认证

### POST /api/auth/login

使用 Dashboard 密码进行认证。

**请求：**
```json
{
  "password": "your-password"
}
```

**响应：**
```json
{
  "ok": true
}
```

设置 session cookie 用于后续请求。

### GET /api/auth/status

检查认证状态。

**响应：**
```json
{
  "authenticated": true,
  "password_set": true
}
```

### POST /api/auth/setup

设置或更改 Dashboard 密码。

**请求：**
```json
{
  "password": "new-password",
  "confirm": "new-password"
}
```

---

## 聊天 API（核心）

这些是第三方客户端最重要的端点。

### POST /api/chat

同步聊天 — 发送消息，等待完整回复。

**请求：**
```json
{
  "message": "你好，你能做什么？",
  "channel": "pico",
  "chat_id": "my-client-user"
}
```

**响应：**
```json
{
  "reply": "你好！我是 PicoClaw，你的 AI 助手...",
  "chat_id": "my-client-user"
}
```

**cURL 示例：**
```bash
curl -X POST http://localhost:18800/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "2+2 等于几？",
    "channel": "pico",
    "chat_id": "test-user"
  }'
```

### POST /api/chat/stream

流式聊天 — 以 Server-Sent Events (SSE) 接收回复。

**请求：** 与 `/api/chat` 相同。

**响应：** SSE 流，事件类型为 `message`：
```
data: {"content": "你好"}
data: {"content": "！"}
data: {"content": "我是"}
data: {"content": "PicoClaw"}
data: {"done": true}
```

**cURL 示例：**
```bash
curl -X POST http://localhost:18800/api/chat/stream \
  -H "Content-Type: application/json" \
  -d '{
    "message": "给我讲个笑话",
    "channel": "pico",
    "chat_id": "test-user"
  }'
```

---

## 会话管理

### GET /api/sessions

列出所有聊天会话。

**查询参数：**
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `offset` | int | 0 | 分页偏移 |
| `limit` | int | 50 | 每页最大结果数 |

**响应：**
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

获取某个会话的完整消息历史。

**响应：**
```json
{
  "id": "session-abc123",
  "messages": [
    {
      "role": "user",
      "content": "你好",
      "timestamp": "2026-06-27T10:00:00Z"
    },
    {
      "role": "assistant",
      "content": "你好！有什么可以帮助你的？",
      "timestamp": "2026-06-27T10:00:05Z"
    }
  ]
}
```

### DELETE /api/sessions/{id}

删除会话及其关联文件。

**响应：**
```json
{
  "ok": true
}
```

---

## 模型管理

### GET /api/models

列出所有已配置的模型及其可用状态。

**响应：**
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

测试模型端点连通性（不保存）。

**请求：**
```json
{
  "provider": "mimo",
  "model": "mimo-v2.5",
  "api_base": "https://api.xiaomimimo.com/v1",
  "api_key": "your-api-key"
}
```

**响应：**
```json
{
  "ok": true,
  "latency_ms": 230,
  "status": "connected"
}
```

---

## 配置管理

### GET /api/config

返回完整系统配置（JSON 格式）。

### PATCH /api/config

使用 JSON Merge Patch (RFC 7396) 部分更新配置。仅修改指定的字段。

**请求：**
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

替换整个配置。请谨慎使用。

### POST /api/config/reset

将配置重置为出厂默认值。保留 API 密钥和安全凭据。

---

## 网关生命周期

### GET /api/gateway/status

检查网关运行状态。

**响应：**
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

启动网关子进程。

### POST /api/gateway/stop

停止运行中的网关。

### POST /api/gateway/restart

重启网关（停止 + 启动）。

---

## 健康检查

### GET /health

返回健康状态。服务器运行时始终返回 200。

**响应：**
```json
{
  "status": "ok",
  "uptime": "2h30m",
  "pid": 12345
}
```

### GET /ready

返回就绪状态。未就绪时返回 503。

### POST /reload

触发配置重载。受可选 bearer token 保护。

---

## 工具：load_video

AI agent 新增的工具，用于加载和分析本地视频文件。

### 工具定义

| 字段 | 值 |
|------|-----|
| 名称 | `load_video` |
| 参数 | `path`（string，必填） |

### 使用方式

当 AI 调用 `load_video(path="video.mp4")` 时：

1. 验证文件路径在工作区内
2. 检测 MIME 类型（必须为 `video/*`）
3. 将文件存入 media store
4. 返回 `media://` 引用

视频随后会自动：
- 被 `resolveMediaRefs` 编码为 `data:video/mp4;base64,...`
- 被 Provider 以 `video_url` 格式发送给模型
- 被支持视频输入的多模态模型（如 MiMo）处理

### 支持的视频格式

MP4、WebM、MOV、AVI、MKV 以及其他 Go `filetype` 库识别的常见格式。

### 大小限制

由配置中的 `agents.defaults.max_media_size` 控制（默认：20MB）。

---

## 多语言 API 文档

| 语言 | 文件 |
|------|------|
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
