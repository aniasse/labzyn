// Package auth provides password hashing and JWT issuance/verification.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidToken is returned when a token cannot be verified.
var ErrInvalidToken = errors.New("invalid token")

// HashPassword returns a bcrypt hash of the password.
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword reports whether password matches the stored hash.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

type claims struct {
	jwt.RegisteredClaims
}

// NewToken issues a signed JWT whose subject is the user id.
func NewToken(secret, userID string, ttl time.Duration) (string, error) {
	now := time.Now()
	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(secret))
}

// ParseToken verifies a token and returns the user id (subject).
func ParseToken(secret, tokenString string) (string, error) {
	tok, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil || !tok.Valid {
		return "", ErrInvalidToken
	}
	c, ok := tok.Claims.(*claims)
	if !ok || c.Subject == "" {
		return "", ErrInvalidToken
	}
	return c.Subject, nil
}
