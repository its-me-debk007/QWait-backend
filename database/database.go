package database

import (
	"log"
	"os"
	"github.com/its-me-debk007/QWait_backend/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dbUrl := os.Getenv("DATABASE_URL")

	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	DB = db
	if err := db.AutoMigrate(
		new(model.User),
		new(model.Store),
		new(model.StoreStats),
	); err != nil {
		log.Fatalln("AUTO_MIGRATION_ERROR")
	}
}
