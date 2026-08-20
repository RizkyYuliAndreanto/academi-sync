# REST API Response Contract & Guidelines

## 1. Envelope Structure

Setiap response REST API pada layanan **Bimbingan Online** dikembalikan dalam bentuk JSON envelope yang bersifat *mutually exclusive* (hanya mengandung satu dari `data` atau `error`), serta selalu mengikutsertakan `request_id` yang konsisten dengan header `X-Request-ID`.

### 1.1 Success Envelope
Dikembalikan untuk semua respons sukses (`200 OK`, `201 Created`).

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Bimbingan Skripsi"
  },
  "meta": {
    "next_cursor": "eyJpZCI6MTIzfQ==",
    "has_more": true
  },
  "request_id": "c91121a6-5208-48ff-9b80-3294a20f1d52"
}
```

*Catatan `meta`:* Field `meta` opsional (`omitempty`) dan hanya hadir pada endpoint berhalaman (pagination) atau menyajikan agregat metadata.

### 1.2 Error Envelope
Dikembalikan untuk semua respons gagal (`4xx` dan `5xx`).

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Data yang dikirimkan tidak valid",
    "fields": {
      "email": "Format email tidak sesuai"
    }
  },
  "request_id": "c91121a6-5208-48ff-9b80-3294a20f1d52"
}
```

### 1.3 No Content (HTTP 204)
Untuk respons `204 No Content`:
- **Tidak ada JSON body** (body kosong 0 bytes).
- Header `X-Request-ID` tetap tersedia pada HTTP response header.

---

## 2. Katalog Error Code & HTTP Status Mapping

Seluruh kode kesalahan publik dikelompokkan dalam enum terpusat (`ErrorCode`). Detail kesalahan database atau stack trace internal **dilarang keras** disajikan dalam `code` maupun `message`.

| Error Code | HTTP Status | Keterangan |
| :--- | :--- | :--- |
| `VALIDATION_ERROR` | `400 Bad Request` | Gagal validasi input (misal format email, field wajib). Mengisi map `fields`. |
| `INVALID_REQUEST` | `400 Bad Request` | Body JSON malformed atau parameter query tidak valid. |
| `UNAUTHENTICATED` | `401 Unauthorized` | Client belum terautentikasi atau session/token tidak valid. |
| `FORBIDDEN` | `403 Forbidden` | Client terautentikasi tetapi tidak memiliki hak akses atas aksi tersebut. |
| `RESOURCE_NOT_FOUND` | `404 Not Found` | Endpoint atau resource yang diminta tidak ditemukan. |
| `METHOD_NOT_ALLOWED` | `405 Method Not Allowed` | Metode HTTP tidak didukung pada endpoint tersebut. |
| `RESOURCE_CONFLICT` | `409 Conflict` | Konflik status resource (misal email sudah terdaftar). |
| `PAYLOAD_TOO_LARGE` | `413 Payload Too Large` | Ukuran payload/file melebihi batas yang diizinkan. |
| `UNSUPPORTED_MEDIA_TYPE` | `415 Unsupported Media Type` | Header Content-Type tidak didukung. |
| `RATE_LIMITED` | `429 Too Many Requests` | Melebihi batas jumlah request (rate limit). |
| `INTERNAL_ERROR` | `500 Internal Server Error` | Kesalahan internal server / panic yang ditangkap. |
| `SERVICE_UNAVAILABLE` | `503 Service Unavailable` | Layanan downstream atau pemeliharaan server. |

---

## 3. Kebijakan Otorisasi & Pencegahan IDOR (401 vs 403 vs 404)

1. **Gunakan `401 Unauthenticated`**:
   - Ketika pengguna belum login, session kadaluarsa, atau token JWT tidak valid.
2. **Gunakan `404 Resource Not Found` (IDOR Prevention)**:
   - Jika pengguna terautentikasi meminta resource berdasar ID (`GET /guidance-sessions/{id}` atau `GET /documents/{id}`) yang **bukan miliknya** atau tidak berkepentingan dengannya.
   - Menyembunyikan keberadaan UUID resource dari pihak yang tidak berhak.
3. **Gunakan `403 Forbidden`**:
   - Jika pengguna mengetahui keberadaan resource (misal sesi bimbingan di mana ia adalah anggota), tetapi melakukan aksi khusus yang tidak diizinkan oleh rolenya (misal mahasiswa mencoba menyelesaikan sesi bimbingan).

---

## 4. Standar Cursor-Based Pagination Metadata

Untuk endpoint koleksi (list data), gunakan cursor-based pagination melalui field `meta`:

```json
{
  "data": [...],
  "meta": {
    "next_cursor": "cursor_string_opaque",
    "has_more": true,
    "total_count": 150
  },
  "request_id": "..."
}
```

---

## 5. Konsistensi Request ID

Seluruh helper response mengambil `request_id` dari Gin context (`X-Request-ID`). Terdapat garansi konsistensi:
- Header `X-Request-ID` pada HTTP response
- Field `request_id` pada JSON envelope
- Field `request_id` pada structured application logger (`log/slog`)
