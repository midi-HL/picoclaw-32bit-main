[English](README.md) | [中文](README.zh.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Português](README.pt-br.md) | [Tiếng Việt](README.vi.md) | [Français](README.fr.md) | [Italiano](README.it.md) | [Bahasa Indonesia](README.id.md) | **Malay**

# PicoClaw Fork — Rujukan API

Dokumen ini menerangkan endpoint API yang tersedia dalam fork ini. Fokus pada endpoint yang berguna untuk pembangunan klien pihak ketiga.

---

## Endpoint API Utama

| Endpoint | Kaedah | Penerangan |
|----------|--------|-----------|
| `/api/chat` | POST | Chat serentak |
| `/api/chat/stream` | POST | Chat penstriman (SSE) |
| `/api/sessions` | GET | Senarai sesi |
| `/api/sessions/{id}` | GET | Sejarah sesi |
| `/api/models` | GET | Senarai model |
| `/api/config` | GET/PATCH | Baca/tulis konfigurasi |
| `/api/gateway/status` | GET | Status gateway |
| `/health` | GET | Health check |

---

Untuk butiran, rujuk versi Bahasa Inggeris: [English API Reference](README.md)
