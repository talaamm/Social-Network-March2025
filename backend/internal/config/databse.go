package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func GetDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./data/social-network.db")
	if err != nil {
		log.Println(err)
	}
	return db
}

func InitDB() *sql.DB {
	if db != nil {
		return db
	}
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		os.Mkdir("./data", os.ModePerm)
	}

	var err error
	db, err = sql.Open("sqlite3", "./data/social-network.db")
	if err != nil {
		log.Fatal("❌ Failed to open database:", err)
	}
	if err := applyMigrations(); err != nil {
		log.Fatal("❌ Failed to apply migrations:", err)
	}

	return db
}

func CloseDB() {
	if db != nil {
		db.Close()
	}
}

func applyMigrations() error {
	migrationDir := "migrations"
	absPath, err := filepath.Abs(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute migration path: %v", err)
	}

	files, err := os.ReadDir(absPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			migrationPath := filepath.Join(absPath, file.Name())
			migrationSQL, err := os.ReadFile(migrationPath)
			if err != nil {
				return fmt.Errorf("failed to read migration %s: %v", file.Name(), err)
			}

			_, err = db.Exec(string(migrationSQL))
			if err != nil {
				return fmt.Errorf("failed to execute migration %s: %v", file.Name(), err)
			}
			fmt.Println("🔹 Applied migration:", file.Name())
		}
	}
	return nil
}
