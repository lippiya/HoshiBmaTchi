package gorm_user

import (
	"fmt"
	"log"
	"os"
	"sync"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var(
	db sync.Once
	instance *gorm.DB
)

func GetDatabase() *gorm.DB{
	db.Do(func(){
		template := "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta"

		host := getEnvOrDefault("USER_POSTGRES_HOST", "localhost")
		user := getEnvOrDefault("USER_POSTGRES_USER", "postgres")
		password := getEnvOrDefault("USER_POSTGRES_PASSWORD", "password")
		dbname := getEnvOrDefault("USER_POSTGRES_DB", "hoshiBmaTchi")
		port := getEnvOrDefault("USER_POSTGRES_PORT", "5432")

		dsn := fmt.Sprintf(
			template,
			host,
			user,
			password,
			dbname,
			port,
		)

		var err error
		instance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil{
			log.Fatalln("Database connection failed")
		}
	
	})
	return instance
}