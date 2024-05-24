package utils

import (
	"testing"
)

func Test_CheckHashedPassword(t *testing.T) {
	pass := "secretpassword"
	hashedPass, err := HashPassword(pass)
	if err != nil {
		t.Error("Error hashing password:", err)
	}

	if !CheckPasswordHash(pass, hashedPass) {
		t.Error("Password and hashed password do not match")
	}
}

func Test_ComparePassword(t *testing.T) {
	pass := "secretpassword"
	hash := "$2a$14$8swPsf5bmdBs5p0rrEGUleOHkW8HkC9nyhWGnM5RFN3UVZWmwWcde"

	if !CheckPasswordHash(pass, hash) {
		t.Error("Passwords do not match")
	}
}
