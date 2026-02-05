package models

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("seedance.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	DB.AutoMigrate(&Project{}, &Storyboard{})
}
