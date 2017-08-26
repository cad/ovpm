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
func SetupDB(dialect string, args ...interface{}) {
	if len(args) > 0 && args[0] == "" {
		args[0] = _DefaultDBPath
	}
	var err error
	db, err = gorm.Open(dialect, args...)
	if err != nil {
		logrus.Fatalf("couldn't open sqlite database %v: %v", args, err)
	}

	db.AutoMigrate(&DBUser{})
	db.AutoMigrate(&DBServer{})
	db.AutoMigrate(&DBRevoked{})
	db.AutoMigrate(&DBNetwork{})
}

// CeaseDB closes the database.
//
// It should be run at the exit of the program.
func CeaseDB() {
	db.Close()
}
