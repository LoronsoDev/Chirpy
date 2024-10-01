package auth

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	new_uuid := uuid.New()
	secret := "secret"
	expiresIn := 15 * time.Second
	tokenString, err := MakeJWT(new_uuid, secret, expiresIn)
	if err != nil {
		t.Errorf("MakeJWT() error = %v", err)
	}
	jwtSecret := os.Getenv("JWT_SECRET")

	id, err := ValidateJWT(tokenString, jwtSecret)

	if id != new_uuid {
		t.Errorf("id was not verified correctly: %v != %v", new_uuid, id)
	}

	if err != nil {
		t.Error(err.Error())
	}

	_, err = ValidateJWT(tokenString, "randomSecret")
	if err == nil {
		t.Error("Token with invalid secret is being validated")
	}

	new_uuid = uuid.New()
	secret = "secret"
	expiresIn = 0
	tokenString, _ = MakeJWT(new_uuid, secret, expiresIn)

	_, err = ValidateJWT(tokenString, jwtSecret)
	if err == nil {
		t.Error("Token is expired but it is still being verified as valid")
	}

}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		headers       http.Header
		expectedToken string
		expectError   bool
	}{
		{
			name: "valid bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer validtoken123"},
			},
			expectedToken: "validtoken123",
			expectError:   false,
		},
		{
			name: "missing authorization header",
			headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			expectedToken: "",
			expectError:   true,
		},
		{
			name: "invalid format in authorization header",
			headers: http.Header{
				"Authorization": []string{"invalidtoken123"},
			},
			expectedToken: "",
			expectError:   true,
		},
		{
			name: "invalid bearer prefix",
			headers: http.Header{
				"Authorization": []string{"Basic validtoken123"},
			},
			expectedToken: "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetBearerToken(tt.headers)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got error: %v", tt.expectError, err)
			}
			if token != tt.expectedToken {
				t.Errorf("expected token: %v, got token: %v", tt.expectedToken, token)
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrongPassword",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Password doesn't match different hash",
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password1,
			hash:     "invalidhash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
