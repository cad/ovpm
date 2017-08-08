package ovpm

import (
	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"

	// We blank import sqlite here because gorm needs it.
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB

// CloseDB closes the database.
func CloseDB() {
	db.Close()
}

func init() {
	var err error
	db, err = gorm.Open("sqlite3", DefaultDBPath)
	if err != nil {
		logrus.Fatalf("couldn't open sqlite database %s: %v", DefaultDBPath, err)
	}

	db.AutoMigrate(&DBUser{})
	db.AutoMigrate(&DBNetwork{})
	db.AutoMigrate(&DBServer{})
	db.AutoMigrate(&DBRevoked{})

}
