package jwt

import (
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	Init("test-secret-key")
	if jwtSecret == nil {
		t.Error("jwtSecret should be initialized")
	}
}

func TestGenerateAndParse(t *testing.T) {
	Init("test-secret-key")

	token, err := Generate(1, true, 0, 72)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	claims, err := Parse(token)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if claims.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", claims.UserID)
	}
	if !claims.IsAdmin {
		t.Error("expected IsAdmin true")
	}
	if claims.TokenVersion != 0 {
		t.Errorf("expected TokenVersion 0, got %d", claims.TokenVersion)
	}
}

func TestParse_InvalidToken(t *testing.T) {
	Init("test-secret-key")

	_, err := Parse("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestParse_WrongSecret(t *testing.T) {
	Init("secret-1")
	token, _ := Generate(1, false, 0, 1)

	Init("secret-2") // different secret
	_, err := Parse(token)
	if err == nil {
		t.Error("expected error for token signed with different secret")
	}
}

func TestGenerate_ExpiredToken(t *testing.T) {
	Init("test-secret-key")
	// Generate with 0 hours = expired immediately
	token, err := Generate(1, false, 0, 0)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	_, err = Parse(token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestTokenVersion(t *testing.T) {
	Init("test-secret-key")

	token, err := Generate(1, false, 5, 72)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	claims, err := Parse(token)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if claims.TokenVersion != 5 {
		t.Errorf("expected TokenVersion 5, got %d", claims.TokenVersion)
	}
}

func TestMultipleTokens(t *testing.T) {
	Init("test-secret-key")

	tokens := make([]string, 10)
	for i := 0; i < 10; i++ {
		token, err := Generate(uint(i+1), i%2 == 0, i, 72)
		if err != nil {
			t.Fatalf("Generate %d error: %v", i, err)
		}
		tokens[i] = token
	}

	for i, token := range tokens {
		claims, err := Parse(token)
		if err != nil {
			t.Fatalf("Parse token %d error: %v", i, err)
		}
		if claims.UserID != uint(i+1) {
			t.Errorf("token %d: expected UserID %d, got %d", i, i+1, claims.UserID)
		}
		if claims.TokenVersion != i {
			t.Errorf("token %d: expected TokenVersion %d, got %d", i, i, claims.TokenVersion)
		}
	}
}