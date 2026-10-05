package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func getEnvOrPanic(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic("variable d'environnement manquante: " + key)
	}
	return val
}

func jwtSecret() []byte {
	return []byte(getEnvOrPanic("JWT_SECRET"))
}

func adminPasswordHash() string {
	return getEnvOrPanic("ADMIN_PASSWORD_HASH")
}

func CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(adminPasswordHash()), []byte(password))
	return err == nil
}

func GenerateToken() (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   "admin",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret())
}

func ValidateToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("méthode de signature inattendue")
		}
		return jwtSecret(), nil
	})

	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token invalide")
	}
	return nil
}
