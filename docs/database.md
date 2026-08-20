# Database Architecture & Migration Standard

Dokumen ini menjelaskan arsitektur PostgreSQL connection pool, manajemen migrasi database, dan skema dasar untuk aplikasi **Bimbingan Online**.

---

## 1. Pilihan Teknologi

- **Driver / Connection Pool**: `github.com/jackc/pgx/v5/pgxpool`
- **Migration Tool**: `github.com/golang-migrate/migrate/v4`
- **Migration Source**: `embed.FS` (`source/iofs`) untuk menyatukan file `.sql` ke dalam binary.
- **Primary Key Standard**: Aplikasi menghasilkan UUID v4 (tidak bergantung pada ekstensi PostgreSQL).

---

## 2. Parameter Konfigurasi Database

Variabel lingkungan yang mendukung koneksi PostgreSQL:

```env
DATABASE_ENABLED=true
DATABASE_URL=postgres://user:password@localhost:5432/bimbingan_db?sslmode=disable
DATABASE_MAX_CONNECTIONS=10
DATABASE_MIN_CONNECTIONS=1
DATABASE_CONNECT_TIMEOUT=5s
DATABASE_MAX_CONN_LIFETIME=30m
DATABASE_MAX_CONN_IDLE_TIME=5m
DATABASE_HEALTH_CHECK_PERIOD=30s
```

### Aturan Lingkungan Production
- Pada `APP_ENV=production`, `DATABASE_ENABLED` **WAJIB** bernilai `true`. Aplikasi akan gagal startup jika `DATABASE_ENABLED=false` pada lingkungan produksi.

---

## 3. Struktur Tabel Baseline & Constraint

1. **`users`**:
   - Menampung akun Mahasiswa, Dosen, dan Admin.
   - Constraint: `LOWER(email)` unique index, `role IN ('dosen', 'mahasiswa', 'admin')`.
2. **`refresh_sessions`**:
   - Menampung token sesi refresh untuk autentikasi.
   - Constraint: Foreign key ke `users(id)` dengan `ON DELETE CASCADE`. Menyimpan `token_hash` (bukan token mentah).
3. **`guidance_sessions`**:
   - Sesi bimbingan antara Dosen dan Mahasiswa.
   - Constraint: `lecturer_id != student_id`, `scheduled_end_at > scheduled_start_at`, `version >= 1`.
4. **`documents`**:
   - File dokumen yang diunggah dalam sesi bimbingan.
   - Constraint: `size_bytes > 0`, `page_count > 0`, `UNIQUE (storage_bucket, storage_key)`.
5. **`annotations`**:
   - Anotasi pada halaman dokumen.
   - Constraint: `page_number > 0`, `document_version > 0`, partial index pada anotasi aktif (`WHERE deleted_at IS NULL`).
6. **`annotation_snapshots`**:
   - Snapshot status anotasi real-time per halaman.
   - Constraint: `UNIQUE (document_id, page_number, state_version)`.

---

## 4. Perintah & Prosedur Migrasi

Migrasi **TIDAK** dijalankan secara otomatis saat server HTTP startup (`cmd/server/main.go`), melainkan melalui CLI terpisah `cmd/migrate/main.go`.

### 4.1 Perintah CLI Migrasi

```bash
# Menjalankan seluruh migrasi yang belum dieksekusi (Up)
go run ./cmd/migrate up

# Memeriksa versi migrasi saat ini dan status dirty
go run ./cmd/migrate version

# Melakukan rollback sejumlah langkah (Down)
go run ./cmd/migrate down 1
```

### 4.2 Prosedur Deployment Production
- Migrasi dieksekusi sebagai langkah awal deployment CI/CD pipeline sebelum kontainer aplikasi baru diaktifkan.
- Apabila terjadi *dirty state* (misal migrasi gagal di tengah jalan), runner akan menolak melanjutkan deployment secara otomatis demi keamanan data.

---

## 5. Health Probe (/health/live & /health/ready)

- **`GET /health/live`**: Mengembalikan status `200 OK` tanpa tergantung pada PostgreSQL untuk mencegah restart loop saat database bermasalah.
- **`GET /health/ready`**: Melakukan ping dengan timeout 2 detik ke PostgreSQL.
  - Jika sehat (atau jika DB di-disable di dev/test): `200 OK` `{"data": {"status": "ready"}, "request_id": "..."}`.
  - Jika koneksi terputus: `503 Service Unavailable` `{"error": {"code": "SERVICE_UNAVAILABLE", "message": "Layanan belum siap"}, "request_id": "..."}`.
