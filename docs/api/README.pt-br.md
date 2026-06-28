[English](README.md) | [中文](README.zh.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Tiếng Việt](README.vi.md) | [Français](README.fr.md) | [Italiano](README.it.md) | [Bahasa Indonesia](README.id.md) | [Malay](README.ms.md) | **Português**

# PicoClaw Fork — Referência da API

Este documento descreve os endpoints da API disponíveis neste fork. Foco nos endpoints úteis para desenvolvimento de clientes de terceiros.

---

## Principais Endpoints da API

| Endpoint | Método | Descrição |
|----------|--------|-----------|
| `/api/chat` | POST | Chat síncrono |
| `/api/chat/stream` | POST | Chat em streaming (SSE) |
| `/api/sessions` | GET | Lista de sessões |
| `/api/sessions/{id}` | GET | Histórico da sessão |
| `/api/models` | GET | Lista de modelos |
| `/api/config` | GET/PATCH | Leitura/escrita de configuração |
| `/api/gateway/status` | GET | Status do gateway |
| `/health` | GET | Health check |

---

Para detalhes, consulte a versão em inglês: [English API Reference](README.md)
