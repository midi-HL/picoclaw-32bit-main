[English](README.md) | [中文](README.zh.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Português](README.pt-br.md) | [Tiếng Việt](README.vi.md) | [Français](README.fr.md) | [Bahasa Indonesia](README.id.md) | [Malay](README.ms.md) | **Italiano**

# PicoClaw Fork — Riferimento API

Questo documento descrive gli endpoint API disponibili in questo fork. Focus sugli endpoint utili per lo sviluppo di client di terze parti.

---

## Principali Endpoint API

| Endpoint | Metodo | Descrizione |
|----------|--------|-------------|
| `/api/chat` | POST | Chat sincrono |
| `/api/chat/stream` | POST | Chat in streaming (SSE) |
| `/api/sessions` | GET | Lista sessioni |
| `/api/sessions/{id}` | GET | Cronologia sessione |
| `/api/models` | GET | Lista modelli |
| `/api/config` | GET/PATCH | Lettura/scrittura configurazione |
| `/api/gateway/status` | GET | Stato gateway |
| `/health` | GET | Health check |

---

Per i dettagli, consultare la versione inglese: [English API Reference](README.md)
