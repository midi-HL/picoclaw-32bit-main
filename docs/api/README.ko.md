[English](README.md) | [中文](README.zh.md) | [日本語](README.ja.md) | [Português](README.pt-br.md) | [Tiếng Việt](README.vi.md) | [Français](README.fr.md) | [Italiano](README.it.md) | [Bahasa Indonesia](README.id.md) | [Malay](README.ms.md) | **한국어**

# PicoClaw Fork — API 레퍼런스

이 문서는 이 fork에서 사용 가능한 API 엔드포인트를 설명합니다. 서드파이티 클라이언트 개발에 유용한 엔드포인트에 중점을 둡니다.

---

## 주요 API 엔드포인트

| 엔드포인트 | 방법 | 설명 |
|-----------|------|------|
| `/api/chat` | POST | 동기 채팅 |
| `/api/chat/stream` | POST | 스트리밍 채팅 (SSE) |
| `/api/sessions` | GET | 세션 목록 |
| `/api/sessions/{id}` | GET | 세션 기록 조회 |
| `/api/models` | GET | 모델 목록 |
| `/api/config` | GET/PATCH | 설정 읽기/쓰기 |
| `/api/gateway/status` | GET | 게이트웨이 상태 |
| `/health` | GET | 헬스체크 |

---

자세한 내용은 영어 버전을 참조하세요: [English API Reference](README.md)
