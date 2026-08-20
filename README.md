# Bimbingan-Online — Monorepo Project

Platform Bimbingan Skripsi/Tugas Akhir Online Real-Time (WebRTC, PDF Synchronization, Live Annotations, and Session Management).

## Project Structure

```text
.
├── AGENTS.md                   # Agent execution instructions & invariants
├── README.md                   # Main project overview
├── ai-agent-project-tasks/     # Governance, architecture, task specifications, and checklists
├── backend/                    # Go (Gin) API & Real-time WebRTC/WebSocket Service
├── frontend/                   # React + TypeScript Client
├── infra/                      # Docker Compose, TURN (coturn), MinIO configurations
├── docs/                       # Architectural Decision Records (ADRs) & documentation
└── scripts/                    # Utility, seed, and deployment scripts
```

## Quick Start (Development)

### Backend

```bash
cd backend
go run ./cmd/server
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```
