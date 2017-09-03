package ovpm

import "testing"

func TestDBSetup(t *testing.T) {
	// Initialize:
	Testing = true

	// Prepare:
	// Test:

	// Create database.
	CreateDB("sqlite3", ":memory:")

	// Is database created?
	if db == nil {
		t.Fatalf("database is expected to be not nil but it's nil")
	}
}

func TestDBCease(t *testing.T) {
	// Initialize:
	Testing = true

	// Prepare:
	CreateDB("sqlite3", ":memory:")
	user := dbUserModel{Username: "testUser"}
	db.Save(&user)

	// Test:
	// Close database.
	db.Cease()

	var users []dbUserModel
	db.Find(&users)

	// Is length zero?
	if len(users) != 0 {
		t.Fatalf("length of user should be 0 but it's %d", len(users))
	}
}
