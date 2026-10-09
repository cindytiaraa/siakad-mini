package dto

// Pointer dipakai agar field yang tidak dikirim bisa dibedakan dari nilai kosong.

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// StudentRequest dipakai POST dan PUT /students.
// Pada PUT, nim (dan email) yang ikut terkirim diabaikan.
type StudentRequest struct {
	NIM      *string  `json:"nim"`
	Nama     *string  `json:"nama"`
	Email    *string  `json:"email"`
	Prodi    *string  `json:"prodi"`
	Angkatan *int     `json:"angkatan"`
	IPK      *float64 `json:"ipk_terakhir"`
}

type EnrollmentRequest struct {
	CourseID      *int64 `json:"course_id"`
	TahunAkademik string `json:"tahun_akademik"`
}
