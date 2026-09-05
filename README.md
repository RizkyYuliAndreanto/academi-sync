<div align="center">

# 🎓 Bimbingan Online

**Platform Bimbingan Skripsi / Tugas Akhir secara Real-Time — seolah duduk bersebelahan dengan dosen pembimbing.**

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react&logoColor=black)](https://react.dev)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?style=flat-square&logo=typescript&logoColor=white)](https://www.typescriptlang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat-square&logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![MinIO](https://img.shields.io/badge/Storage-MinIO-C72E49?style=flat-square&logo=minio&logoColor=white)](https://min.io)
[![WebRTC](https://img.shields.io/badge/Video-WebRTC-333333?style=flat-square&logo=webrtc&logoColor=white)](https://webrtc.org)

_Video call · Sinkronisasi PDF · Anotasi live · Manajemen sesi_

</div>

---

## ✨ Fitur

|     | Fitur                   | Deskripsi                                                                                       |
| --- | ----------------------- | ----------------------------------------------------------------------------------------------- |
| 🎥  | **Video Call WebRTC**   | Konferensi 1-on-1 mahasiswa ↔ dosen dengan TURN (coturn) untuk traversal NAT                    |
| 📄  | **Sinkronisasi PDF**    | Dokumen aktif & halaman aktif dikendalikan server — scroll dosen, ikut mahasiswa                |
| ✏️  | **Anotasi Live**        | Gambar & catatan di atas PDF (Konva.js), koordinat _normalized_ agar tampil sama di semua layar |
| 🔌  | **Real-Time Signaling** | WebSocket dengan satu writer goroutine per koneksi dan bounded queue                            |
| 📁  | **Manajemen Dokumen**   | Upload PDF ke MinIO (bucket private, akses via presigned URL)                                   |
| 🔐  | **Auth & Session**      | JWT via HttpOnly cookie — tidak pernah disimpan di Local Storage                                |
| 📋  | **Manajemen Sesi**      | Jadwal bimbingan, riwayat, dan status sesi per mahasiswa                                        |

## 🏗️ Arsitektur

```text
┌─────────────┐   HTTPS/WS    ┌──────────────────┐      ┌────────────┐
│  Frontend   │◄─────────────►│  Backend (Go/Gin)│◄────►│ PostgreSQL │
│ React + TS  │  REST + WS    │  API + Signaling │      └────────────┘
│  + Konva    │               │                  │      ┌────────────┐
└──────┬──────┘               │   ┌──────────┐   │◄────►│   MinIO    │
       │                      │   │ Signaling│   │      │  (private) │
       │      P2P (DTLS/SRTP) │   └────┬─────┘   │      └────────────┘
       └────────◄─────────────┼────────┘
            media & data      │  TURN credentials
                      ┌───────▼───────┐
                      │ coturn (TURN) │
                      └───────────────┘
```

**Prinsip utama:** server _authoritative_. Klien tidak pernah dipercaya untuk `sender_id`, role, ownership, page count, atau permission — semuanya divalidasi dan diurutkan oleh server.

## 📂 Struktur Proyek

```text
.
├── backend/                    # Go (Gin) — API + WebSocket/WebRTC signaling
│   ├── cmd/                    # Entrypoint (server, migrate)
│   ├── internal/               # app, config, database, middleware, health, ...
│   ├── migrations/             # SQL migration (golang-migrate, embedded)
│   └── tests/
├── frontend/                   # React + TypeScript + Vite + Tailwind
├── ai-agent-project-tasks/     # Governance: baseline, task spec, checklist, ADR
│   ├── 00-governance/ … 10-testing-release/
│   ├── ARCHITECTURE_BASELINE.md
│   ├── EVENT_CATALOG.md
│   └── SECURITY_GATES.md · TEST_MATRIX.md · MASTER_CHECKLIST.md
├── docs/                       # ADR & dokumentasi (api/, database.md)
├── infra/                      # Docker Compose, coturn, MinIO
└── scripts/                    # Utilitas, seed, deployment
```

## 🚀 Quick Start

**Prasyarat:** Go ≥ 1.25, Node.js ≥ 18, (opsional) Docker untuk PostgreSQL/MinIO.

```bash
# 1. Konfigurasi environment
cp .env.example .env

# 2. Backend
cd backend
go run ./cmd/server          # → http://localhost:8080

# 3. Frontend (terminal baru)
cd frontend
npm install
npm run dev                  # → http://localhost:5173
```

Database dan storage _diaktifkan bertahap_ lewat flag di `.env`
(`DATABASE_ENABLED`, `STORAGE_ENABLED`, `TURN_ENABLED`) sesuai fase task yang sedang dikerjakan.

## ⚙️ Konfigurasi Utama

| Variabel               | Default                 | Keterangan                                 |
| ---------------------- | ----------------------- | ------------------------------------------ |
| `HTTP_PORT`            | `8080`                  | Port API backend                           |
| `HTTP_ALLOWED_ORIGINS` | `http://localhost:5173` | CORS origin (pisah dengan koma)            |
| `DATABASE_ENABLED`     | `false`                 | Aktifkan koneksi PostgreSQL (pgxpool)      |
| `DATABASE_URL`         | —                       | DSN PostgreSQL                             |
| `STORAGE_ENABLED`      | `false`                 | Aktifkan MinIO untuk dokumen PDF           |
| `TURN_ENABLED`         | `false`                 | Aktifkan TURN untuk WebRTC lintas jaringan |

Lengkap di [`.env.example`](.env.example) — log, timeout HTTP, dan kebijakan cookie juga bisa diatur di sana.

---

<div align="center">
<sub>Dibangun dengan Go · Gin · Gorilla WebSocket · React · TypeScript · Konva · PostgreSQL · MinIO · coturn</sub>
</div>
