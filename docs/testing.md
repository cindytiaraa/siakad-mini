# Testing

1. Unit test: `go test ./...` (batas SKS, validator, rate limiter).
2. Smoke test status code: `./scripts/smoke.sh` (200, 201, 204, 401, 403, 404, 409, 422, 429).
3. Skenario manual yang perlu dicoba sendiri:
   - 500: matikan database saat server hidup lalu panggil `GET /api/v1/courses` dengan `APP_ENV=production`; response harus generik.
   - Kuota penuh: ambil `IF401` (kuota 2) dengan 3 mahasiswa berbeda; yang ketiga mendapat 422.
   - Batas SKS: mahasiswa ber-IPK 2,49 (NIM `187221000008`) mengambil mata kuliah sampai melewati 18 SKS pada satu tahun akademik; 422 dengan pesan sisa SKS.
   - Soft delete: `DELETE /students/{id}`, lalu pastikan hilang dari list, detail 404, login 401.
   - Konkurensi: kirim dua `POST /enrollments` paralel untuk course berkuota sisa 1; satu 201, satu 422.
