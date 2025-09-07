package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"shared/db"
)

func main() {
	var (
		action = flag.String("action", "migrate", "Action to perform: migrate, check")
		path   = flag.String("path", "../../shared/db", "Path to migrations directory")
	)
	flag.Parse()

	switch *action {
	case "check":
		if err := db.CheckConnection(); err != nil {
			log.Fatalf("❌ Database connection failed: %v", err)
		}
		
	case "migrate":
		fmt.Println("🚀 Running database migrations...")
		
		migrationsPath, err := filepath.Abs(*path)
		if err != nil {
			log.Fatalf("❌ Failed to resolve migrations path: %v", err)
		}

		dbURL := db.GetConnectionURL()
		if err := db.RunMigrations(dbURL, migrationsPath); err != nil {
			log.Fatalf("❌ Migration failed: %v", err)
		}

		fmt.Println("✅ All migrations completed successfully!")
		
	default:
		fmt.Printf("❌ Unknown action: %s\n", *action)
		fmt.Println("Available actions: check, migrate")
		os.Exit(1)
	}
}