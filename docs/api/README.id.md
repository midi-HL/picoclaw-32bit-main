[English](README.md) | [中文](README.zh.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Português](README.pt-br.md) | [Tiếng Việt](README.vi.md) | [Français](README.fr.md) | [Italiano](README.it.md) | [Malay](README.ms.md) | **Bahasa Indonesia**

# PicoClaw Fork — Referensi API

Dokumen ini menjelaskan endpoint API yang tersedia dalam fork ini. Fokus pada endpoint yang berguna untuk pengembangan klien pihak ketiga.

---

## Endpoint API Utama

| Endpoint | Metode | Deskripsi |
|----------|--------|-----------|
| `/api/chat` | POST | Chat sinkron |
| `/api/chat/stream` | POST | Chat streaming (SSE) |
| `/api/sessions` | GET | Daftar sesi |
| `/api/sessions/{id}` | GET | Riwayat sesi |
| `/api/models` | GET | Daftar model |
| `/api/config` | GET/PATCH | Baca/tulis konfigurasi |
| `/api/gateway/status` | GET | Status gateway |
| `/health` | GET | Health check |

---

Untuk detail, lihat versi Bahasa Inggris: [English API Reference](README.md)
