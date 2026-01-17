package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJwtToken(t *testing.T) {
	userId := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	tokenSecret := "my_secret_key"
	expiresIn := 2 * time.Hour

	token, err := MakeJWT(userId, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("Error creating JWT: %v", err)
	}

	returnedUserId, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("Error validating JWT: %v", err)
	}

	if returnedUserId != userId {
		t.Fatalf("Expected userId %s, got %s", userId, returnedUserId)
	}
}
