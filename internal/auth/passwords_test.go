package auth

import (
	"testing"
)

func TestPasswordHashing(t *testing.T) {
	password := "my_secure_password"

	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}

	match, err := CheckPasswordHash(password, hashedPassword)
	if err != nil {
		t.Fatalf("Error checking password: %v", err)
	}
	if !match {
		t.Fatalf("Expected password to match hashed password")
	}

	wrongPassword := "wrong_password"
	match, err = CheckPasswordHash(wrongPassword, hashedPassword)
	if err != nil {
		t.Fatalf("Error checking wrong password: %v", err)
	}
	if match {
		t.Fatalf("Expected wrong password not to match hashed password")
	}
}