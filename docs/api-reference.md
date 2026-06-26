# PicoClaw API 接口文档

## 基础信息

- **基础 URL**: `http://{host}:{port}`
- **默认端口**: `18800`（launcher）
- **认证**: 无需认证（内网部署模式）
- **格式**: 请求/响应均为 JSON，流式端点使用 SSE

---

## 端点总览

| 方法 | 路径 | 状态 | 说明 |
|------|------|:----:|------|
| POST | `/api/chat` | 新增 | 发送消息，同步获取完整回复 |
| POST | `/api/chat/stream` | 新增 | 发送消息，SSE 流式获取回复 |
| GET | `/api/sessions` | 现有 | 列出所有会话 |
| GET | `/api/sessions/{id}` | 现有 | 获取会话消息历史 |
| DELETE | `/api/sessions/{id}` | 现有 | 删除会话 |
| GET | `/api/models` | 现有 | 列出所有已配置模型 |
| POST | `/api/models/default` | 现有 | 设置默认模型 |
| POST | `/api/models/{index}/test` | 现有 | 测试模型连接 |
| GET | `/api/config` | 现有 | 读取完整系统配置 |
| PATCH | `/api/config` | 现有 | 增量更新配置 |
| GET | `/api/gateway/status` | 现有 | 网关运行状态 |
| GET | `/api/tools` | 现有 | 可用工具列表 |
| GET | `/api/channels/catalog` | 现有 | 频道目录 |
| GET | `/api/system/version` | 现有 | 系统版本 |

---

## 聊天端点（新增）

### POST /api/chat

发送一条消息给 Agent，阻塞等待完整回复后返回 JSON。

**请求体**:

```json
{
  "message": "请用 Python 写一个快速排序",
  "session_id": "abc123",
  "model": "openai/gpt-4o"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|:----:|------|
| `message` | string | 是 | 用户消息文本 |
| `session_id` | string | 否 | 已有会话 ID，不传则自动创建新会话 |
| `model` | string | 否 | 临时覆盖当前模型，不传则用默认模型 |

**响应体**:

```json
{
  "session_id": "abc123",
  "model": "openai/gpt-4o",
  "reply": "以下是快速排序的 Python 实现：\n\ndef quicksort(arr):\n    ...",
  "tool_calls": [
    {
      "name": "exec",
      "args": "{\"command\":\"python --version\"}",
      "result": "Python 3.12.0"
    }
  ],
  "timestamp": "2026-06-26T10:30:00Z"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `session_id` | string | 当前会话 ID |
| `model` | string | 实际使用的模型名 |
| `reply` | string | AI 的完整回复文本 |
| `tool_calls` | array | 过程中调用的工具及结果 |
| `timestamp` | string | 回复时间 (RFC3339) |

**错误响应**:

```json
{"error": "gateway not available"}
{"error": "message is required"}
{"error": "request timed out waiting for agent response"}
```

**curl 示例**:

```bash
curl -s -X POST http://localhost:18800/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"今天天气怎么样"}'
```

---

### POST /api/chat/stream

同 `/api/chat`，但以 SSE 格式逐块推送回复内容。

**请求体**: 与 `/api/chat` 一致。

**响应** (`Content-Type: text/event-stream`):

```
event: start
data: {"session_id":"abc123"}

event: text
data: "以下"

event: text
data: "是快速"

event: text
data: "排序的"

event: done
data: {"session_id":"abc123","reply":"以下是快速排序的 Python 实现：..."}
```

**事件类型**:

| 事件 | 说明 |
|------|------|
| `start` | 流开始，包含 session_id |
| `text` | 文本片段 |
| `tool_call` | 工具调用信息 |
| `done` | 流结束，包含完整回复 |

**curl 示例**:

```bash
curl -N -X POST http://localhost:18800/api/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"message":"讲个笑话"}'
```

---

## 会话管理（现有）

### GET /api/sessions

列出所有会话摘要。

**查询参数**:
- `offset` (int, 默认 0): 分页偏移
- `limit` (int, 默认 20): 每页数量

**响应**: 会话摘要数组

```bash
curl -s http://localhost:18800/api/sessions
```

### GET /api/sessions/{id}

获取指定会话的完整消息历史。

```bash
curl -s http://localhost:18800/api/sessions/abc123
```

### DELETE /api/sessions/{id}

删除指定会话。

```bash
curl -s -X DELETE http://localhost:18800/api/sessions/abc123
```

---

## 模型管理（现有）

### GET /api/models

列出所有已配置的模型及其状态。

```bash
curl -s http://localhost:18800/api/models
```

### POST /api/models/default

设置默认模型（切换模型）。

```bash
curl -s -X POST http://localhost:18800/api/models/default \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4o-mini"}'
```

### POST /api/models/{index}/test

测试指定模型的连接。

```bash
curl -s -X POST http://localhost:18800/api/models/0/test
```

---

## 配置管理（现有）

### GET /api/config

读取完整系统配置。

```bash
curl -s http://localhost:18800/api/config
```

### PATCH /api/config

增量更新配置（JSON Merge Patch）。

```bash
curl -s -X PATCH http://localhost:18800/api/config \
  -H "Content-Type: application/json" \
  -d '{
    "voice": {
      "model_name": "openai/whisper-1",
      "tts_model_name": "openai/tts-1"
    },
    "agents": {
      "defaults": {
        "image_model": "openai/gpt-4o",
        "max_media_size": 104857600
      }
    }
  }'
```

---

## 系统信息（现有）

### GET /api/gateway/status

网关运行状态。

```bash
curl -s http://localhost:18800/api/gateway/status
```

### GET /api/tools

可用 Agent 工具列表。

```bash
curl -s http://localhost:18800/api/tools
```

### GET /api/channels/catalog

频道目录。

```bash
curl -s http://localhost:18800/api/channels/catalog
```

### GET /api/system/version

系统版本信息。

```bash
curl -s http://localhost:18800/api/system/version
```

---

## Python 调用示例

```python
import requests

BASE = "http://192.168.1.100:18800"

# 基础对话
r = requests.post(f"{BASE}/api/chat", json={
    "message": "请用 Python 写一个快速排序"
})
reply = r.json()
print(f"回复: {reply['reply']}")

# 在已有会话中继续对话
r = requests.post(f"{BASE}/api/chat", json={
    "message": "能给这个排序加上注释吗？",
    "session_id": reply["session_id"]
})
print(r.json()["reply"])

# 流式对话
import sseclient  # pip install sseclient-py

r = requests.post(f"{BASE}/api/chat/stream", json={
    "message": "讲一个关于程序员的冷笑话"
}, stream=True)

client = sseclient.SSEClient(r)
for event in client.events():
    if event.event == "text":
        print(event.data, end="", flush=True)
    elif event.event == "done":
        print("\n--- 回复结束 ---")

# 会话管理
sessions = requests.get(f"{BASE}/api/sessions").json()

# 模型切换
requests.post(f"{BASE}/api/models/default", json={"model": "openai/gpt-4o-mini"})

# 配置修改
requests.patch(f"{BASE}/api/config", json={
    "voice": {"tts_model_name": "openai/tts-1"}
})
```
