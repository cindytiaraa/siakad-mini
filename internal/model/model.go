package model

import "time"

const (
	RoleAdmin     = "admin"
	RoleMahasiswa = "mahasiswa"
)

type User struct {
	ID       int64
	Email    string
	Password string
	Role     string
}

// Student: JSON mengikuti contoh response PDF (id, nim, nama, prodi, angkatan, ipk_terakhir).
type Student struct {
	ID       int64    `json:"id"`
	UserID   int64    `json:"-"`
	NIM      string   `json:"nim"`
	Nama     string   `json:"nama"`
	Prodi    string   `json:"prodi"`
	Angkatan int      `json:"angkatan"`
	IPK      *float64 `json:"ipk_terakhir"`
}

type Course struct {
	ID        int64  `json:"id"`
	KodeMK    string `json:"kode_mk"`
	NamaMK    string `json:"nama_mk"`
	SKS       int    `json:"sks"`
	Semester  int    `json:"semester"`
	Kuota     int    `json:"kuota"`
	Terisi    int    `json:"terisi"`
	SisaKuota int    `json:"sisa_kuota"`
}

type Enrollment struct {
	ID            int64     `json:"id"`
	StudentID     int64     `json:"student_id"`
	CourseID      int64     `json:"course_id"`
	TahunAkademik string    `json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`
}

// EnrolledCourse adalah satu baris KRS pada detail mahasiswa.
type EnrolledCourse struct {
	EnrollmentID  int64  `json:"enrollment_id"`
	CourseID      int64  `json:"course_id"`
	KodeMK        string `json:"kode_mk"`
	NamaMK        string `json:"nama_mk"`
	SKS           int    `json:"sks"`
	Semester      int    `json:"semester"`
	TahunAkademik string `json:"tahun_akademik"`
}
