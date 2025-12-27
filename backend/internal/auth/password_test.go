package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Fatal("Hash should not be empty")
	}

	if hash == password {
		t.Fatal("Hash should not equal plain password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"
	wrongPassword := "wrongpassword"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Test correct password
	if err := CheckPassword(hash, password); err != nil {
		t.Errorf("CheckPassword should succeed with correct password: %v", err)
	}

	// Test wrong password
	if err := CheckPassword(hash, wrongPassword); err == nil {
		t.Error("CheckPassword should fail with wrong password")
	}
}

func TestHashPasswordDeterminism(t *testing.T) {
	password := "testpassword123"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("First HashPassword failed: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Second HashPassword failed: %v", err)
	}

	// Bcrypt includes random salt, so hashes should differ
	if hash1 == hash2 {
		t.Error("Two hashes of same password should differ (bcrypt uses random salt)")
	}

	// But both should verify against the same password
	if err := CheckPassword(hash1, password); err != nil {
		t.Errorf("First hash should verify: %v", err)
	}

	if err := CheckPassword(hash2, password); err != nil {
		t.Errorf("Second hash should verify: %v", err)
	}
}
