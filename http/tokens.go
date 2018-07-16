package http

import (
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/upframe/api"
)

func tokensGet(w http.ResponseWriter, r *http.Request, c *api.Config) (int, interface{}, error) {
	password := r.FormValue("password")
	if password != api.Password {
		return http.StatusUnauthorized, nil, nil
	}

	// Expires the token and cookie in 24 hour
	expireToken := time.Now().Add(time.Hour * 24).Unix()

	// We'll manually assign the claims but in production you'd insert values from a database
	claims := jwt.StandardClaims{
		ExpiresAt: expireToken,
		Issuer:    "api.upframe.co",
	}

	// Create the token using your claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Signs the token with a secret.
	signedToken, _ := token.SignedString(api.JWTSecret)
	return 200, signedToken, nil
}
