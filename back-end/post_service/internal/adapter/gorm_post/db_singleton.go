package gorm_post

import (
	"fmt"
	"log"
	"os"
	"sync"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var(
	db sync.Once
	instance *gorm.DB
)

func GetDatabase() *gorm.DB{
	db.Do(func(){
		template := "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta"

		dsn := fmt.Sprintf(
			template,
			os.Getenv("POST_POSTGRES_HOST"),
			os.Getenv("POST_POSTGRES_USER"),
			os.Getenv("POST_POSTGRES_PASSWORD"),
			os.Getenv("POST_POSTGRES_DB"),
			os.Getenv("POST_POSTGRES_PORT"),

		)

		var err error
		instance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil{
			log.Fatalln("Post database connection failed")
		}



	})
	return instance
}
