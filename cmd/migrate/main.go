package main

import (
	"fmt"
	"log"
	"os"

	"siakad-mini/internal/config"
	"siakad-mini/internal/database"
)

const usage = `Pemakaian:
  go run ./cmd/migrate up         jalankan semua migration baru
  go run ./cmd/migrate down       batalkan migration terakhir
  go run ./cmd/migrate down all   batalkan semua migration`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	cfg := config.Load()
	cfg.RequireDatabase()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("gagal terhubung ke database: %v", err)
	}
	defer db.Close()

	switch os.Args[1] {
	case "up":
		err = database.MigrateUp(db, cfg.MigrationsDir)
	case "down":
		all := len(os.Args) > 2 && os.Args[2] == "all"
		err = database.MigrateDown(db, cfg.MigrationsDir, all)
	default:
		fmt.Println(usage)
		os.Exit(1)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("selesai")
}
