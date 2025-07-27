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

func ValidateToken(tokenString string, secret string) (bool, error) {

	//we just want to test if we can parse the token with the server secret.
	//TODO: Validate time
	//Note: Claims should be validate elsewhere
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return false, err
	}

	return true, nil
}
