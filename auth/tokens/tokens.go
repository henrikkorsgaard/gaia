package tokens

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	_ "github.com/joho/godotenv/autoload"
)

type UserToken struct {
	Scope string `json:"scope"` //e.g. crm, api,
	jwt.RegisteredClaims
}

func NewUserToken(userId string, secret string) (string, error) {

	rc := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "Gaia",
		Subject:   userId,
		Audience:  []string{"crm", "data", "invoice"},
	}

	claims := UserToken{
		"crm:write data:read invoice:read", rc,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}
