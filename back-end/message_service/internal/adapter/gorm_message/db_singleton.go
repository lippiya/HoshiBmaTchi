package gorm_message

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
			os.Getenv("MESSAGE_POSTGRES_HOST"),
			os.Getenv("MESSAGE_POSTGRES_USER"),
			os.Getenv("MESSAGE_POSTGRES_PASSWORD"),
			os.Getenv("MESSAGE_POSTGRES_DB"),
			os.Getenv("MESSAGE_POSTGRES_PORT"),
		)

		var err error
		instance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil{
			log.Fatalln("Message database connection failed")
		}
	})
	return instance
}
