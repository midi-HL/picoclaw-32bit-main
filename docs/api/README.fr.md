[English](README.md) | [中文](README.zh.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Português](README.pt-br.md) | [Tiếng Việt](README.vi.md) | [Italiano](README.it.md) | [Bahasa Indonesia](README.id.md) | [Malay](README.ms.md) | **Français**

# PicoClaw Fork — Référence API

Ce document décrit les endpoints API disponibles dans ce fork. Focus sur les endpoints utiles au développement de clients tiers.

---

## Principaux Endpoints API

| Endpoint | Méthode | Description |
|----------|---------|-------------|
| `/api/chat` | POST | Chat synchrone |
| `/api/chat/stream` | POST | Chat en streaming (SSE) |
| `/api/sessions` | GET | Liste des sessions |
| `/api/sessions/{id}` | GET | Historique de session |
| `/api/models` | GET | Liste des modèles |
| `/api/config` | GET/PATCH | Lecture/écriture de configuration |
| `/api/gateway/status` | GET | Statut du gateway |
| `/health` | GET | Health check |

---

Pour les détails, consultez la version anglaise: [English API Reference](README.md)
