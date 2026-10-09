package service

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"siakad-mini/internal/apperror"
	"siakad-mini/internal/dto"
	"siakad-mini/internal/model"
	"siakad-mini/internal/repository"
	"siakad-mini/internal/validator"
)

type StudentService struct {
	db          *sql.DB
	users       *repository.UserRepository
	students    *repository.StudentRepository
	enrollments *repository.EnrollmentRepository
}

func NewStudentService(db *sql.DB, users *repository.UserRepository, students *repository.StudentRepository, enrollments *repository.EnrollmentRepository) *StudentService {
	return &StudentService{db: db, users: users, students: students, enrollments: enrollments}
}

func trim(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}

func validateAngkatan(a *int, errs validator.Errors) {
	switch {
	case a == nil:
		errs.Add("angkatan", "Angkatan wajib diisi")
	case *a < 1000 || *a > 9999:
		errs.Add("angkatan", "Angkatan harus 4 digit")
	case *a > time.Now().Year():
		errs.Add("angkatan", "Angkatan tidak boleh melebihi tahun berjalan")
	}
}

func validateIPK(ipk *float64, errs validator.Errors) {
	if ipk != nil && (*ipk < 0 || *ipk > 4) {
		errs.Add("ipk_terakhir", "IPK harus antara 0,00 dan 4,00")
	}
}

// mapUniqueViolation: pelanggaran UNIQUE yang lolos pre-check (race) tetap dijawab 422.
func mapUniqueViolation(err error) error {
	constraint, ok := repository.UniqueViolation(err)
	if !ok {
		return err
	}
	switch constraint {
	case "users_email_key":
		return apperror.Validation(map[string][]string{"email": {"Email sudah terdaftar"}})
	case "students_nim_key":
		return apperror.Validation(map[string][]string{"nim": {"NIM sudah terdaftar"}})
	}
	return err
}

func (s *StudentService) List(ctx context.Context, f repository.StudentFilter) ([]model.Student, int, error) {
	return s.students.List(ctx, s.db, f)
}

// Create membuat users (role mahasiswa, password awal = hash(nim)) dan students dalam satu transaction.
func (s *StudentService) Create(ctx context.Context, req dto.StudentRequest) (*model.Student, error) {
	nim := trim(req.NIM)
	nama := trim(req.Nama)
	email := strings.ToLower(trim(req.Email))
	prodi := trim(req.Prodi)

	errs := validator.Errors{}
	if nim == "" {
		errs.Add("nim", "NIM wajib diisi")
	} else if !validator.IsNIM(nim) {
		errs.Add("nim", "NIM harus 12 digit angka")
	}
	if nama == "" {
		errs.Add("nama", "Nama wajib diisi")
	}
	if email == "" {
		errs.Add("email", "Email wajib diisi")
	} else if !validator.IsEmail(email) {
		errs.Add("email", "Format email tidak valid")
	}
	if prodi == "" {
		errs.Add("prodi", "Prodi wajib diisi")
	}
	validateAngkatan(req.Angkatan, errs)
	validateIPK(req.IPK, errs)

	// Pre-check keunikan hanya untuk field yang formatnya sudah valid.
	if _, bad := errs["nim"]; !bad {
		exists, err := s.students.NIMExists(ctx, s.db, nim)
		if err != nil {
			return nil, err
		}
		if exists {
			errs.Add("nim", "NIM sudah terdaftar")
		}
	}
	if _, bad := errs["email"]; !bad {
		exists, err := s.users.EmailExists(ctx, s.db, email)
		if err != nil {
			return nil, err
		}
		if exists {
			errs.Add("email", "Email sudah terdaftar")
		}
	}
	if errs.Any() {
		return nil, apperror.Validation(errs)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(nim), bcrypt.DefaultCost) // password awal = nim
	if err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	userID, err := s.users.Create(ctx, tx, email, string(hash), model.RoleMahasiswa)
	if err != nil {
		return nil, mapUniqueViolation(err)
	}
	st := &model.Student{
		UserID:   userID,
		NIM:      nim,
		Nama:     nama,
		Prodi:    prodi,
		Angkatan: *req.Angkatan,
		IPK:      req.IPK,
	}
	if err := s.students.Create(ctx, tx, st); err != nil {
		return nil, mapUniqueViolation(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return st, nil
}

type StudentDetail struct {
	model.Student
	MataKuliah []model.EnrolledCourse `json:"mata_kuliah"`
	TotalSKS   int                    `json:"total_sks"`
	BatasSKS   int                    `json:"batas_sks"`
}

// Detail: admin boleh melihat siapa saja; mahasiswa hanya data dirinya sendiri (403 jika bukan).
// total_sks dihitung pada tahun akademik terbaru yang memiliki enrollment.
func (s *StudentService) Detail(ctx context.Context, role string, callerStudentID, id int64) (*StudentDetail, error) {
	if role == model.RoleMahasiswa && id != callerStudentID {
		return nil, apperror.Forbidden("Anda tidak memiliki akses ke resource ini")
	}

	st, err := s.students.FindActiveByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if st == nil {
		return nil, apperror.NotFound("Data mahasiswa tidak ditemukan")
	}

	courses, err := s.enrollments.ListByStudent(ctx, s.db, st.ID)
	if err != nil {
		return nil, err
	}

	total := 0
	tahun, err := s.enrollments.LatestTahun(ctx, s.db, st.ID)
	if err != nil {
		return nil, err
	}
	if tahun != "" {
		total, err = s.enrollments.SumSKS(ctx, s.db, st.ID, tahun)
		if err != nil {
			return nil, err
		}
	}

	return &StudentDetail{Student: *st, MataKuliah: courses, TotalSKS: total, BatasSKS: BatasSKS(st.IPK)}, nil
}

// Update memperbarui nama, prodi, angkatan, ipk_terakhir. nim tidak dapat diubah.
// Urutan: 404 dahulu, baru validasi (422).
func (s *StudentService) Update(ctx context.Context, id int64, req dto.StudentRequest) (*model.Student, error) {
	existing, err := s.students.FindActiveByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperror.NotFound("Data mahasiswa tidak ditemukan")
	}

	nama := trim(req.Nama)
	prodi := trim(req.Prodi)

	errs := validator.Errors{}
	if nama == "" {
		errs.Add("nama", "Nama wajib diisi")
	}
	if prodi == "" {
		errs.Add("prodi", "Prodi wajib diisi")
	}
	validateAngkatan(req.Angkatan, errs)
	validateIPK(req.IPK, errs)
	if errs.Any() {
		return nil, apperror.Validation(errs)
	}

	st, err := s.students.Update(ctx, s.db, id, nama, prodi, *req.Angkatan, req.IPK)
	if err != nil {
		return nil, err
	}
	if st == nil {
		return nil, apperror.NotFound("Data mahasiswa tidak ditemukan")
	}
	return st, nil
}

func (s *StudentService) Delete(ctx context.Context, id int64) error {
	deleted, err := s.students.SoftDelete(ctx, s.db, id)
	if err != nil {
		return err
	}
	if !deleted {
		return apperror.NotFound("Data mahasiswa tidak ditemukan")
	}
	return nil
}
