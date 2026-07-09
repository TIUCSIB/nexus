package crypto

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("test-password-123")
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	if hash == "" {
		t.Fatal("hash should not be empty")
	}
	if hash == "test-password-123" {
		t.Fatal("hash should not equal plaintext password")
	}
}

func TestCheckPassword_Correct(t *testing.T) {
	password := "my-secure-password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}

	if !CheckPassword(password, hash) {
		t.Error("CheckPassword should return true for correct password")
	}
}

func TestCheckPassword_Wrong(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}

	if CheckPassword("wrong-password", hash) {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestCheckPassword_Empty(t *testing.T) {
	hash, err := HashPassword("some-password")
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}

	if CheckPassword("", hash) {
		t.Error("CheckPassword should return false for empty password")
	}
}

func TestHashPassword_Unique(t *testing.T) {
	// Same password should produce different hashes (bcrypt uses random salt)
	password := "same-password"
	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	if hash1 == hash2 {
		t.Error("bcrypt hashes should be unique due to random salt")
	}

	// Both should still validate
	if !CheckPassword(password, hash1) {
		t.Error("hash1 should validate")
	}
	if !CheckPassword(password, hash2) {
		t.Error("hash2 should validate")
	}
}

func TestHashPassword_UTF8(t *testing.T) {
	password := "密码123!@#"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword error for UTF-8 password: %v", err)
	}

	if !CheckPassword(password, hash) {
		t.Error("UTF-8 password should validate correctly")
	}
}