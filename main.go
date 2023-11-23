package main

import (
	"brick-code-test/app/repository"
	"brick-code-test/app/usecase"
	"brick-code-test/model"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)



func main() {
	dsn := "host=localhost user=postgres password=password dbname=scrapper port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&model.Product{})
	productRepository := repository.New(db)
	doneChan := make(chan struct{})
	productUsecase := usecase.New(productRepository, doneChan)
	productUsecase.RunScheduledTasks()
	
}