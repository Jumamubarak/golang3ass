package data

import (
	"database/sql"
	_ "github.com/lib/pq"
	"testing"
)

func Test_SaveUser(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://postgres:nurik05@localhost:5432/aitugolang?sslmode=disable")
	if err != nil {
		t.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	sqlDb := UserInfoModel{db: db}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	user := &UserInfo{
		Fname:        "John",
		Sname:        "Doe",
		Email:        "john@example.com",
		PasswordHash: "hashed_password",
		UserRole:     "user",
		Activated:    false,
		Version:      1,
	}
	savedUser, err := sqlDb.CreateUser(user)
	if err != nil {
		t.Fatalf("Error saving user: %v", err)
	}

	if savedUser.Fname != user.Fname || savedUser.Email != user.Email {
		t.Errorf("Retrieved user does not match expected values")
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Error rolling back transaction: %v", err)
	}
}

func Test_GetByID(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://postgres:nurik05@localhost:5432/aitugolang?sslmode=disable")
	if err != nil {
		t.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	sqlDb := UserInfoModel{db: db}

	// Create a test user
	user := &UserInfo{
		Fname:        "Test",
		Sname:        "User",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		UserRole:     "user",
		Activated:    false,
		Version:      1,
	}
	createdUser, err := sqlDb.CreateUser(user)
	if err != nil {
		t.Fatalf("Error creating test user: %v", err)
	}
	defer sqlDb.deleteByID(createdUser.ID)

	// Test GetByID
	retrievedUser, err := sqlDb.GetByID(createdUser.ID)
	if err != nil {
		t.Fatalf("Error retrieving user by ID: %v", err)
	}

	// Verify retrieved user matches created user
	if retrievedUser.ID != createdUser.ID || retrievedUser.Email != createdUser.Email {
		t.Errorf("Retrieved user does not match expected values")
	}
}

func Test_DeleteByID(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://postgres:nurik05@localhost:5432/aitugolang?sslmode=disable")
	if err != nil {
		t.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	sqlDb := UserInfoModel{db: db}

	// Create a test user
	user := &UserInfo{
		Fname:        "Test",
		Sname:        "User",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		UserRole:     "user",
		Activated:    false,
		Version:      1,
	}
	createdUser, err := sqlDb.CreateUser(user)
	if err != nil {
		t.Fatalf("Error creating test user: %v", err)
	}
	defer sqlDb.deleteByID(createdUser.ID)

	err = sqlDb.deleteByID(createdUser.ID)
	if err != nil {
		t.Fatalf("Error deleting test user: %v", err)
	}
}

func Test_UpdateByID(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://postgres:nurik05@localhost:5432/aitugolang?sslmode=disable")
	if err != nil {
		t.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	sqlDb := UserInfoModel{db: db}

	user := &UserInfo{
		Fname:        "Test",
		Sname:        "User",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		UserRole:     "user",
		Activated:    false,
		Version:      1,
	}
	createdUser, err := sqlDb.CreateUser(user)
	if err != nil {
		t.Fatalf("Error creating test user: %v", err)
	}
	defer sqlDb.deleteByID(createdUser.ID)

	updatedUser := &UserInfo{
		ID:        createdUser.ID,
		Fname:     "Updated",
		Sname:     "User",
		Email:     "updated@example.com",
		UserRole:  "admin",
		Activated: true,
		Version:   2,
	}

	updatedUserInfo, err := sqlDb.updateByID(createdUser.ID, updatedUser)
	if err != nil {
		t.Fatalf("Error updating user by ID: %v", err)
	}

	if updatedUserInfo == nil {
		t.Error("Updated user info is nil")
	} else {
		if updatedUserInfo.Fname != updatedUser.Fname || updatedUserInfo.Email != updatedUser.Email {
			t.Errorf("Updated user info does not match expected values")
		}
	}
}
