package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"

	"siakad-mini/internal/config"
	"siakad-mini/internal/database"
	"siakad-mini/internal/model"
)

const (
	adminEmail    = "admin@siakad.test"
	adminPassword = "Admin12345"
)

var names = []string{
	"Rina Putri", "Budi Santoso", "Siti Aisyah", "Agus Prasetyo", "Dewi Lestari",
	"Eko Wijaya", "Fitri Handayani", "Gilang Ramadhan", "Hana Pertiwi", "Indra Kusuma",
	"Jihan Safitri", "Kevin Hartono", "Laras Ayu", "Mario Setiawan", "Nadia Rahma",
	"Oscar Pratama", "Putri Anggraini", "Rizky Maulana", "Sinta Dewi", "Tegar Nugroho",
}

// Sengaja menyebar di ketiga tingkat batas SKS (>=3.00, 2.50-2.99, <2.50) termasuk nilai batas.
var ipks = []float64{
	3.45, 3.80, 3.20, 3.00, 2.99, 2.75, 2.50, 2.49, 2.10, 3.65,
	3.10, 2.85, 2.60, 2.30, 1.95, 3.90, 3.35, 2.70, 2.45, 3.05,
}

var prodis = []string{"Sistem Informasi", "Teknik Informatika", "Manajemen Informatika"}

type course struct {
	kode, nama     string
	sks, sem, kuota int
}

var courses = []course{
	{"IF101", "Algoritma dan Pemrograman", 4, 1, 40},
	{"IF102", "Matematika Diskrit", 3, 1, 40},
	{"IF103", "Pengantar Teknologi Informasi", 2, 1, 50},
	{"IF201", "Struktur Data", 4, 3, 35},
	{"IF202", "Basis Data", 4, 3, 35},
	{"IF203", "Jaringan Komputer", 3, 3, 30},
	{"IF301", "Rekayasa Perangkat Lunak", 3, 5, 30},
	{"IF302", "Pemrograman Web", 3, 5, 30},
	{"IF303", "Pemrograman Berbasis Enterprise", 3, 6, 25},
	{"IF401", "Kecerdasan Buatan", 3, 7, 2}, // kuota kecil untuk menguji "kuota penuh"
}

func main() {
	cfg := config.Load()
	cfg.RequireDatabase()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("gagal terhubung ke database: %v", err)
	}
	defer db.Close()

	if err := seed(context.Background(), db); err != nil {
		log.Fatalf("seeding gagal: %v", err)
	}
}

func seed(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1 admin
	if _, created, err := ensureUser(ctx, tx, adminEmail, adminPassword, model.RoleAdmin); err != nil {
		return err
	} else if created {
		fmt.Println("admin dibuat :", adminEmail)
	}

	// 20 mahasiswa
	newStudents := 0
	for i, nama := range names {
		nim := fmt.Sprintf("187221%06d", i+1)
		email := nim + "@student.siakad.test"

		userID, created, err := ensureUser(ctx, tx, email, nim, model.RoleMahasiswa) // password awal = nim
		if err != nil {
			return err
		}
		if !created {
			continue
		}
		_, err = tx.ExecContext(ctx,
			`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
			 VALUES ($1, $2, $3, $4, $5, $6::numeric)
			 ON CONFLICT (nim) DO NOTHING`,
			userID, nim, nama, prodis[i%len(prodis)], 2022+(i%3), ipks[i])
		if err != nil {
			return err
		}
		newStudents++
	}
	fmt.Println("mahasiswa baru:", newStudents)

	// 10 mata kuliah
	newCourses := 0
	for _, c := range courses {
		res, err := tx.ExecContext(ctx,
			`INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (kode_mk) DO NOTHING`,
			c.kode, c.nama, c.sks, c.sem, c.kuota)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			newCourses++
		}
	}
	fmt.Println("mata kuliah baru:", newCourses)

	if err := tx.Commit(); err != nil {
		return err
	}
	fmt.Println("seeding selesai")
	return nil
}

// ensureUser membuat user bila belum ada (password disimpan sebagai hash bcrypt).
func ensureUser(ctx context.Context, tx *sql.Tx, email, plain, role string) (id int64, created bool, err error) {
	err = tx.QueryRowContext(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&id)
	if err == nil {
		return id, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return 0, false, err
	}
	err = tx.QueryRowContext(ctx,
		`INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id`,
		email, string(hash), role).Scan(&id)
	return id, err == nil, err
}
