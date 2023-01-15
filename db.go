package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init() *gorm.DB {
	db, err := gorm.Open(postgres.Open(config.DBUrl), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		log.Fatalln(err)
	}
	db.AutoMigrate(
		// &Account{}, &Profile{}, &Token{},
		// &Product{},
		&Import{}, &ImportDetail{},
	)
	return db
}
