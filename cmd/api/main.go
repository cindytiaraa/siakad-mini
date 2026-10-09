package main

import (
	"log"

	"siakad-mini/internal/config"
	"siakad-mini/internal/database"
	"siakad-mini/internal/router"
)

func main() {
	cfg := config.Load()
	cfg.RequireDatabase()
	cfg.RequireJWT()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("gagal terhubung ke database: %v", err)
	}
	defer db.Close()

	r := router.New(cfg, db)
	log.Printf("SIAKAD Mini berjalan di :%s (env=%s)", cfg.Port, cfg.AppEnv)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
