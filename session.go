package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

const (
	cookieName = "jwt_token"
)

func setTokenCookie(w http.ResponseWriter, tokenString string) {
	expiration := time.Now().Add(time.Hour * 24)
	cookie := http.Cookie{Name: cookieName, Value: tokenString, Expires: expiration}
	http.SetCookie(w, &cookie)
}

func createToken(username string) (string, error) {
	expirationTime := time.Now().Add(time.Hour * 24)
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		Subject:   username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(appConfig.Login.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func parseToken(tokenString string) (*jwt.StandardClaims, error) {
	claims := &jwt.StandardClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(appConfig.Login.Secret), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, fmt.Errorf("invalid token")
		}
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
