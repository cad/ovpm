package ovpm

import (
	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"

	// We blank import sqlite here because gorm needs it.
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB

// SetupDB prepares database for use.
//
// It should be run at the start of the program.
func SetupDB() {
	var err error
	db, err = gorm.Open("sqlite3", _DefaultDBPath)
	if err != nil {
		logrus.Fatalf("couldn't open sqlite database %s: %v", _DefaultDBPath, err)
	}

	db.AutoMigrate(&DBUser{})
	db.AutoMigrate(&DBNetwork{})
	db.AutoMigrate(&DBServer{})
	db.AutoMigrate(&DBRevoked{})
}

// CeaseDB closes the database.
//
// It should be run at the exit of the program.
func CeaseDB() {
	db.Close()
}
