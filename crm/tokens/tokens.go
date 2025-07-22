package tokens

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"henrikkorsgaard.dk/gaia/crm/database"

	_ "github.com/joho/godotenv/autoload"
)

type UserToken struct {
	Scope string `json:"scope"` //e.g. crm, api,
	jwt.RegisteredClaims
}

func NewUserToken(user database.User) (string, error) {

	var aud []string
	var scope string
	if user.GaiaId != "" {
		aud = []string{"crm", "data", "invoice"}
		scope = "crm:write data:read invoice:read"
	} else {
		aud = []string{"crm"}
		scope = "crm:match"
	}

	rc := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "Gaia",
		Subject:   user.GaiaId,
		Audience:  aud,
	}

	claims := UserToken{
		scope, rc,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := []byte("tokensecret")
	return token.SignedString(secret)
}
