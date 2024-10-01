package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword -
func HashPassword(password string) (string, error) {
	dat, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(dat), nil
}

// CheckPasswordHash -
func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	expiresAt := time.Now().Add(expiresIn)

	// Create the token with the specified claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(expiresAt), // Set expiration time
		Subject:   userID.String(),
	})
	jwtSecret := os.Getenv("JWT_SECRET")

	return token.SignedString([]byte(jwtSecret))
}

func MakeRefreshToken() (string, error) {
	nBytes := 32
	inputBytes := make([]byte, nBytes)
	_, err := rand.Read(inputBytes)
	if err != nil {
		fmt.Println("error:", err)
		return "", err
	}
	randString := hex.EncodeToString(inputBytes)

	return randString, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	type Claims struct {
		jwt.RegisteredClaims
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return uuid.Nil, errors.New("invalid claims")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header not found")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header not found")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "ApiKey" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}
