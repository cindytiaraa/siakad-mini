package repository

import (
	"context"
	"database/sql"
	"errors"

	"siakad-mini/internal/model"
)

type EnrollmentRepository struct{}

func NewEnrollmentRepository() *EnrollmentRepository { return &EnrollmentRepository{} }

func (r *EnrollmentRepository) Exists(ctx context.Context, q DBTX, studentID, courseID int64, tahun string) (bool, error) {
	var exists bool
	err := q.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM enrollments WHERE student_id = $1 AND course_id = $2 AND tahun_akademik = $3)`,
		studentID, courseID, tahun).Scan(&exists)
	return exists, err
}

func (r *EnrollmentRepository) CountByCourse(ctx context.Context, q DBTX, courseID int64) (int, error) {
	var n int
	err := q.QueryRowContext(ctx, `SELECT COUNT(*) FROM enrollments WHERE course_id = $1`, courseID).Scan(&n)
	return n, err
}

// SumSKS menjumlahkan SKS mahasiswa pada satu tahun akademik.
func (r *EnrollmentRepository) SumSKS(ctx context.Context, q DBTX, studentID int64, tahun string) (int, error) {
	var total int
	err := q.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(c.sks), 0)
		 FROM enrollments e JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1 AND e.tahun_akademik = $2`,
		studentID, tahun).Scan(&total)
	return total, err
}

// LatestTahun mengembalikan tahun akademik terbaru yang punya enrollment ("" jika belum ada).
func (r *EnrollmentRepository) LatestTahun(ctx context.Context, q DBTX, studentID int64) (string, error) {
	var tahun string
	err := q.QueryRowContext(ctx,
		`SELECT tahun_akademik FROM enrollments WHERE student_id = $1 ORDER BY tahun_akademik DESC LIMIT 1`,
		studentID).Scan(&tahun)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return tahun, err
}

func (r *EnrollmentRepository) ListByStudent(ctx context.Context, q DBTX, studentID int64) ([]model.EnrolledCourse, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT e.id, e.course_id, c.kode_mk, c.nama_mk, c.sks, c.semester, e.tahun_akademik
		 FROM enrollments e JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1
		 ORDER BY e.tahun_akademik, e.id`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.EnrolledCourse, 0)
	for rows.Next() {
		var it model.EnrolledCourse
		if err := rows.Scan(&it.EnrollmentID, &it.CourseID, &it.KodeMK, &it.NamaMK, &it.SKS, &it.Semester, &it.TahunAkademik); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *EnrollmentRepository) Create(ctx context.Context, q DBTX, studentID, courseID int64, tahun string) (*model.Enrollment, error) {
	var e model.Enrollment
	err := q.QueryRowContext(ctx,
		`INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		 VALUES ($1, $2, $3)
		 RETURNING id, student_id, course_id, tahun_akademik, created_at`,
		studentID, courseID, tahun).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EnrollmentRepository) FindByID(ctx context.Context, q DBTX, id int64) (*model.Enrollment, error) {
	var e model.Enrollment
	err := q.QueryRowContext(ctx,
		`SELECT id, student_id, course_id, tahun_akademik, created_at FROM enrollments WHERE id = $1`,
		id).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EnrollmentRepository) Delete(ctx context.Context, q DBTX, id int64) error {
	_, err := q.ExecContext(ctx, `DELETE FROM enrollments WHERE id = $1`, id)
	return err
}
