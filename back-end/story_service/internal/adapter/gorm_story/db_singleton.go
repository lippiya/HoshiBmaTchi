package gorm_story

import (
	"fmt"
	"log"
	"os"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

func GetDB() *gorm.DB {
	once.Do(func() {
		host := os.Getenv("STORY_POSTGRES_HOST")
		if host == "" {
			host = "postgres-story"
		}
		user := os.Getenv("STORY_POSTGRES_USER")
		if user == "" {
			user = "postgres"
		}
		password := os.Getenv("STORY_POSTGRES_PASSWORD")
		if password == "" {
			password = "postgres"
		}
		dbname := os.Getenv("STORY_POSTGRES_DB")
		if dbname == "" {
			dbname = "story_db"
		}
		port := os.Getenv("STORY_POSTGRES_PORT")
		if port == "" {
			port = "5432"
		}

		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbname, port)
		var err error
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		if err := MigrateStoryTables(db); err != nil {
			log.Fatalf("Failed to migrate tables: %v", err)
		}

		log.Println("Database connected and migrated successfully")
	})
	return db
}
