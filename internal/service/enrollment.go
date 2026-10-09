package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"siakad-mini/internal/apperror"
	"siakad-mini/internal/dto"
	"siakad-mini/internal/model"
	"siakad-mini/internal/repository"
	"siakad-mini/internal/validator"
)

type EnrollmentService struct {
	db          *sql.DB
	students    *repository.StudentRepository
	courses     *repository.CourseRepository
	enrollments *repository.EnrollmentRepository
}

func NewEnrollmentService(db *sql.DB, students *repository.StudentRepository, courses *repository.CourseRepository, enrollments *repository.EnrollmentRepository) *EnrollmentService {
	return &EnrollmentService{db: db, students: students, courses: courses, enrollments: enrollments}
}

// Create menambah mata kuliah ke KRS. Seluruh pengecekan dan insert berjalan dalam satu transaction:
//  1. kunci baris mahasiswa (mencegah dua request paralel melewati batas SKS)
//  2. kunci baris course (SELECT ... FOR UPDATE) agar kuota tidak terlampaui saat berebut
//  3. cek duplikat (409) -> cek kuota (422) -> cek batas SKS (422)
//  4. insert; unique constraint menjadi pengaman terakhir untuk duplikat (409)
func (s *EnrollmentService) Create(ctx context.Context, studentID int64, req dto.EnrollmentRequest) (*model.Enrollment, error) {
	tahun := strings.TrimSpace(req.TahunAkademik)

	errs := validator.Errors{}
	if req.CourseID == nil || *req.CourseID < 1 {
		errs.Add("course_id", "Course ID wajib diisi")
	}
	if tahun == "" {
		errs.Add("tahun_akademik", "Tahun akademik wajib diisi")
	} else if !validator.IsTahunAkademik(tahun) {
		errs.Add("tahun_akademik", "Format tahun akademik harus seperti 2026/2027-Ganjil")
	}
	if errs.Any() {
		return nil, apperror.Validation(errs)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // no-op bila sudah commit

	student, err := s.students.FindActiveByIDForUpdate(ctx, tx, studentID)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, apperror.Unauthorized("Token tidak valid atau tidak ditemukan")
	}

	course, err := s.courses.FindByIDForUpdate(ctx, tx, *req.CourseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, apperror.Validation(map[string][]string{"course_id": {"Mata kuliah tidak ditemukan"}})
	}

	duplicate, err := s.enrollments.Exists(ctx, tx, student.ID, course.ID, tahun)
	if err != nil {
		return nil, err
	}
	if duplicate {
		return nil, apperror.Conflict("Mata kuliah sudah diambil pada tahun akademik ini")
	}

	terisi, err := s.enrollments.CountByCourse(ctx, tx, course.ID)
	if err != nil {
		return nil, err
	}
	if terisi >= course.Kuota {
		msg := "Kuota mata kuliah sudah penuh"
		return nil, apperror.Unprocessable(msg, map[string][]string{"course_id": {msg}})
	}

	totalSKS, err := s.enrollments.SumSKS(ctx, tx, student.ID, tahun)
	if err != nil {
		return nil, err
	}
	batas := BatasSKS(student.IPK)
	if totalSKS+course.SKS > batas {
		sisa := batas - totalSKS
		if sisa < 0 {
			sisa = 0
		}
		msg := fmt.Sprintf(
			"Total SKS melebihi batas. Batas %d SKS, sudah diambil %d SKS, sisa %d SKS (mata kuliah ini %d SKS)",
			batas, totalSKS, sisa, course.SKS)
		return nil, apperror.Unprocessable(msg, map[string][]string{"course_id": {msg}})
	}

	enrollment, err := s.enrollments.Create(ctx, tx, student.ID, course.ID, tahun)
	if err != nil {
		if _, ok := repository.UniqueViolation(err); ok {
			return nil, apperror.Conflict("Mata kuliah sudah diambil pada tahun akademik ini")
		}
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return enrollment, nil
}

// Delete membatalkan mata kuliah dari KRS milik mahasiswa yang login.
// Kuota otomatis kembali karena `terisi` selalu dihitung dari tabel enrollments.
func (s *EnrollmentService) Delete(ctx context.Context, studentID, id int64) error {
	e, err := s.enrollments.FindByID(ctx, s.db, id)
	if err != nil {
		return err
	}
	if e == nil {
		return apperror.NotFound("Data KRS tidak ditemukan")
	}
	if e.StudentID != studentID {
		return apperror.Forbidden("Anda tidak memiliki akses ke resource ini")
	}
	return s.enrollments.Delete(ctx, s.db, id)
}
