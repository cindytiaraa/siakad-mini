package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"siakad-mini/internal/model"
)

const studentCols = `id, user_id, nim, nama, prodi, angkatan, ipk_terakhir::float8`

type StudentRepository struct{}

func NewStudentRepository() *StudentRepository { return &StudentRepository{} }

type StudentFilter struct {
	Page     int
	PerPage  int
	Prodi    string
	Angkatan *int
	Search   string
	Sort     string // "", "nama", "-ipk_terakhir"
}

type rowScanner interface{ Scan(dest ...any) error }

func scanStudent(row rowScanner) (*model.Student, error) {
	var s model.Student
	var ipk sql.NullFloat64
	if err := row.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &ipk); err != nil {
		return nil, err
	}
	if ipk.Valid {
		v := ipk.Float64
		s.IPK = &v
	}
	return &s, nil
}

func (r *StudentRepository) one(ctx context.Context, q DBTX, query string, args ...any) (*model.Student, error) {
	s, err := scanStudent(q.QueryRowContext(ctx, query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

// FindActiveByID hanya mengembalikan mahasiswa yang belum di-soft delete.
func (r *StudentRepository) FindActiveByID(ctx context.Context, q DBTX, id int64) (*model.Student, error) {
	return r.one(ctx, q, `SELECT `+studentCols+` FROM students WHERE id = $1 AND deleted_at IS NULL`, id)
}

// FindActiveByIDForUpdate mengunci baris mahasiswa sampai transaction selesai.
func (r *StudentRepository) FindActiveByIDForUpdate(ctx context.Context, q DBTX, id int64) (*model.Student, error) {
	return r.one(ctx, q, `SELECT `+studentCols+` FROM students WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, id)
}

func (r *StudentRepository) FindActiveByUserID(ctx context.Context, q DBTX, userID int64) (*model.Student, error) {
	return r.one(ctx, q, `SELECT `+studentCols+` FROM students WHERE user_id = $1 AND deleted_at IS NULL`, userID)
}

// NIMExists juga memeriksa baris yang sudah di-soft delete (constraint UNIQUE tetap berlaku).
func (r *StudentRepository) NIMExists(ctx context.Context, q DBTX, nim string) (bool, error) {
	var exists bool
	err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM students WHERE nim = $1)`, nim).Scan(&exists)
	return exists, err
}

func (r *StudentRepository) Create(ctx context.Context, q DBTX, s *model.Student) error {
	var ipk sql.NullFloat64
	err := q.QueryRowContext(ctx,
		`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		 VALUES ($1, $2, $3, $4, $5, $6::numeric)
		 RETURNING id, ipk_terakhir::float8`,
		s.UserID, s.NIM, s.Nama, s.Prodi, s.Angkatan, s.IPK).Scan(&s.ID, &ipk)
	if err != nil {
		return err
	}
	s.IPK = nil
	if ipk.Valid {
		v := ipk.Float64
		s.IPK = &v
	}
	return nil
}

// Update tidak menyentuh nim. ipk nil = nilai lama dipertahankan.
func (r *StudentRepository) Update(ctx context.Context, q DBTX, id int64, nama, prodi string, angkatan int, ipk *float64) (*model.Student, error) {
	return r.one(ctx, q,
		`UPDATE students
		 SET nama = $2, prodi = $3, angkatan = $4,
		     ipk_terakhir = COALESCE($5::numeric, ipk_terakhir),
		     updated_at = now()
		 WHERE id = $1 AND deleted_at IS NULL
		 RETURNING `+studentCols,
		id, nama, prodi, angkatan, ipk)
}

// SoftDelete mengembalikan false bila tidak ada baris aktif dengan id tersebut.
func (r *StudentRepository) SoftDelete(ctx context.Context, q DBTX, id int64) (bool, error) {
	res, err := q.ExecContext(ctx,
		`UPDATE students SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *StudentRepository) List(ctx context.Context, q DBTX, f StudentFilter) ([]model.Student, int, error) {
	where := []string{"deleted_at IS NULL"}
	args := []any{}
	add := func(cond string, v any) {
		args = append(args, v)
		where = append(where, strings.ReplaceAll(cond, "?", "$"+strconv.Itoa(len(args))))
	}

	if f.Prodi != "" {
		add("LOWER(prodi) = LOWER(?)", f.Prodi)
	}
	if f.Angkatan != nil {
		add("angkatan = ?", *f.Angkatan)
	}
	if f.Search != "" {
		add("(nim ILIKE ? OR nama ILIKE ?)", "%"+escapeLike(f.Search)+"%")
	}
	whereSQL := " WHERE " + strings.Join(where, " AND ")

	var total int
	if err := q.QueryRowContext(ctx, `SELECT COUNT(*) FROM students`+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	order := "id ASC"
	switch f.Sort {
	case "nama":
		order = "nama ASC, id ASC"
	case "-ipk_terakhir":
		order = "ipk_terakhir DESC NULLS LAST, id ASC"
	}

	limitIdx := len(args) + 1
	args = append(args, f.PerPage, (f.Page-1)*f.PerPage)
	query := fmt.Sprintf(`SELECT %s FROM students%s ORDER BY %s LIMIT $%d OFFSET $%d`,
		studentCols, whereSQL, order, limitIdx, limitIdx+1)

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.Student, 0)
	for rows.Next() {
		s, err := scanStudent(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *s)
	}
	return items, total, rows.Err()
}
