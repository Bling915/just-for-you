package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

type TokenManager struct{ secret []byte }

func NewTokenManager(secret []byte) TokenManager {
	return TokenManager{secret: append([]byte(nil), secret...)}
}
func (m TokenManager) IssueAccess(userID int64, now time.Time) (string, time.Time, error) {
	expires := now.Add(15 * time.Minute)
	claims := jwt.RegisteredClaims{Subject: strconv.FormatInt(userID, 10), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(expires)}
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	return s, expires, err
}
func (m TokenManager) ParseAccess(raw string) (int64, error) {
	token, err := jwt.ParseWithClaims(raw, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid access token")
	}
	claims := token.Claims.(*jwt.RegisteredClaims)
	return strconv.ParseInt(claims.Subject, 10, 64)
}
