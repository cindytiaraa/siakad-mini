package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"siakad-mini/internal/model"
)

type CourseRepository struct{}

func NewCourseRepository() *CourseRepository { return &CourseRepository{} }

type CourseFilter struct {
	Semester  *int
	Search    string
	Available bool
}

// List menghitung terisi dari tabel enrollments (LEFT JOIN agar mata kuliah kosong tetap muncul).
func (r *CourseRepository) List(ctx context.Context, q DBTX, f CourseFilter) ([]model.Course, error) {
	where := []string{}
	args := []any{}

	if f.Semester != nil {
		args = append(args, *f.Semester)
		where = append(where, "c.semester = $"+strconv.Itoa(len(args)))
	}
	if f.Search != "" {
		args = append(args, "%"+escapeLike(f.Search)+"%")
		n := strconv.Itoa(len(args))
		where = append(where, "(c.kode_mk ILIKE $"+n+" OR c.nama_mk ILIKE $"+n+")")
	}

	query := `SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, COUNT(e.id)
		FROM courses c LEFT JOIN enrollments e ON e.course_id = c.id`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " GROUP BY c.id"
	if f.Available {
		query += " HAVING COUNT(e.id) < c.kuota"
	}
	query += " ORDER BY c.semester, c.kode_mk"

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Course, 0)
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi); err != nil {
			return nil, err
		}
		c.SisaKuota = c.Kuota - c.Terisi
		if c.SisaKuota < 0 {
			c.SisaKuota = 0
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

// FindByIDForUpdate mengunci baris course (row locking) sampai transaction selesai.
// Dua request yang berebut kuota course yang sama akan diproses bergantian.
func (r *CourseRepository) FindByIDForUpdate(ctx context.Context, q DBTX, id int64) (*model.Course, error) {
	var c model.Course
	err := q.QueryRowContext(ctx,
		`SELECT id, kode_mk, nama_mk, sks, semester, kuota FROM courses WHERE id = $1 FOR UPDATE`,
		id).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
