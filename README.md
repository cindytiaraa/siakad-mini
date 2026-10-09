# SIAKAD Mini — RESTful API (UTS PBE)

Back end layanan akademik sederhana: data mahasiswa, mata kuliah, dan KRS.
Stack: **Go 1.22+**, **Gin**, **PostgreSQL** (pgx), JWT (HS256), bcrypt.

> Framework tidak disebut spesifik di soal ("sesuai praktikum"), sehingga dipilih Gin.
> Jika praktikum memakai framework lain, lapisan `handler`, `middleware`, dan `router` yang perlu disesuaikan;
> `service`, `repository`, dan migration tidak bergantung pada framework.

## Menjalankan

```bash
cp .env.example .env              # sesuaikan DATABASE_URL dan JWT_SECRET
docker compose up -d              # (opsional) PostgreSQL lokal
go mod tidy                       # unduh/rapikan dependensi
go run ./cmd/migrate up           # buat tabel
go run ./cmd/seed                 # 1 admin, 20 mahasiswa, 10 mata kuliah
go run ./cmd/api                  # server di :8080
```

Perintah lain: `go run ./cmd/migrate down` (batalkan 1 migration), `go run ./cmd/migrate down all`,
`go test ./...` (unit test), `./scripts/smoke.sh` (uji status code via curl, server harus hidup).

## Akun seeder

| Role | Email | Password |
|---|---|---|
| admin | `admin@siakad.test` | `Admin12345` |
| mahasiswa (20 akun) | `<nim>@student.siakad.test` | `<nim>` |

NIM seeder: `187221000001` s.d. `187221000020` (mahasiswa pertama: Rina Putri, IPK 3,45).
IPK seeder sengaja menyebar di ketiga tingkat batas SKS (termasuk 3,00 / 2,99 / 2,50 / 2,49).
Mata kuliah `IF401` berkuota 2 untuk menguji "kuota penuh".

## 10 Endpoint (`/api/v1`)

| No | Method | Endpoint | Akses | Sukses |
|---|---|---|---|---|
| 1 | POST | `/auth/login` | Publik | 200 |
| 2 | GET | `/auth/me` | Semua role | 200 |
| 3 | GET | `/students` | Admin | 200 |
| 4 | POST | `/students` | Admin | 201 |
| 5 | GET | `/students/{id}` | Admin, mahasiswa (data sendiri) | 200 |
| 6 | PUT | `/students/{id}` | Admin | 200 |
| 7 | DELETE | `/students/{id}` | Admin | 204 |
| 8 | GET | `/courses` | Semua role | 200 |
| 9 | POST | `/enrollments` | Mahasiswa | 201 |
| 10 | DELETE | `/enrollments/{id}` | Mahasiswa (milik sendiri) | 204 |

Semua endpoint kecuali login memakai header `Authorization: Bearer <token>`.
Status code: 200, 201, 204, 401, 403, 404, 409, 422, 429, 500 (500 selalu generik, stack trace hanya di log server).

Contoh:

```bash
TOKEN=$(curl -s -X POST localhost:8080/api/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"admin@siakad.test","password":"Admin12345"}' | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')
curl -H "Authorization: Bearer $TOKEN" "localhost:8080/api/v1/students?per_page=5&sort=-ipk_terakhir"
```

## Struktur

```
cmd/{api,migrate,seed}   entry point
internal/config          .env
internal/database        koneksi + runner migration
internal/model           struct tabel
internal/dto             request body
internal/validator       aturan validasi (nim, email, tahun akademik)
internal/repository      query SQL (menerima *sql.DB atau *sql.Tx)
internal/service         business rule + transaction
internal/handler         HTTP handler
internal/middleware      auth (authentication), role (authorization), rate limit login, recovery
internal/router          pendaftaran 10 route
migrations               SQL up/down
```

## Business rule dan konkurensi

- Batas SKS: IPK ≥ 3,00 → 24; 2,50–2,99 → 21; < 2,50 → 18 (`internal/service/sks.go`).
- `POST /enrollments` berjalan dalam **satu transaction**: kunci baris mahasiswa, kunci baris course
  (`SELECT … FOR UPDATE`), cek duplikat (409) → kuota (422) → batas SKS (422, pesan memuat sisa SKS) → insert.
  Unique constraint `(student_id, course_id, tahun_akademik)` menjadi pengaman terakhir (409).
- `terisi` dan `sisa_kuota` dihitung dari tabel `enrollments`, sehingga pembatalan KRS otomatis mengembalikan kuota.
- Identitas mahasiswa selalu diambil dari token. KRS milik orang lain → 403.

## Hal yang tidak ditentukan soal dan keputusan yang dipakai

| Topik | Keputusan |
|---|---|
| Token | JWT HS256, masa berlaku `JWT_EXPIRES_IN` (default 3600 detik) |
| Batas SKS "per semester" | Dihitung per `tahun_akademik` yang sama dengan request |
| `terisi` pada `/courses` | Seluruh enrollment pada course tersebut (semua tahun akademik) |
| `total_sks` pada detail mahasiswa | Pada tahun akademik terbaru yang punya enrollment; tiap item `mata_kuliah` memuat `tahun_akademik` dan `enrollment_id` |
| `ipk_terakhir` kosong (NULL) | Dianggap < 2,50 (batas 18 SKS). Pada PUT, bila tidak dikirim nilai lama dipertahankan |
| Mahasiswa membuka `/students/{id}` milik orang lain | Selalu 403 (tanpa membedakan ada/tidaknya data) |
| Soft delete | Token mahasiswa yang sudah dihapus ditolak 401; NIM/email tetap terkunci; enrollment lama dibiarkan |
| Rate limit login | 5 kegagalan (401) per IP+email dalam 1 menit; percobaan berikutnya 429. In-memory, reset saat login berhasil |
| Query tidak valid (`per_page` > 50, `sort` tidak dikenal, JSON rusak) | 422 |
| `tahun_akademik` | `YYYY/YYYY+1-Ganjil` atau `-Genap` |
| PUT `/students/{id}` | `nama`, `prodi`, `angkatan` wajib; `nim` dan `email` di body diabaikan |
| Course tidak ditemukan pada POST `/enrollments` | 422 pada field `course_id` |
| Admin mengelola mata kuliah | Tidak ada endpoint untuk itu di daftar 10 endpoint; mata kuliah hanya dari seeder |
