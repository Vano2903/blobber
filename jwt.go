package main

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type CustomClaims struct {
	Username string `json:"username, omitempty"`
	UserID   int    `json:"ID, omitempty"`
	jwt.StandardClaims
}

func NewCustomClaims(username string, userID int, expiration int64) CustomClaims {
	token := CustomClaims{
		Username: username,
		UserID:   userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiration,
			Issuer:    "vano-jwt-teachers",
		},
	}
	return token
}

func NewSignedToken(claim CustomClaims) (string, error) {
	//unsigned token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	//sign the token
	return token.SignedString([]byte(conf.Secret))
}

func NewJWT(username string, userID int, expiration int64) (string, error) {
	claims := NewCustomClaims(username, userID, expiration)
	return NewSignedToken(claims)
}

func ParseToken(t string) (CustomClaims, error) {
	token, err := jwt.ParseWithClaims(
		t,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(conf.Secret), nil
		},
	)
	if err != nil {
		return CustomClaims{}, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return CustomClaims{}, errors.New("can't parse claims")
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return CustomClaims{}, errors.New("jwt is expired")
	}
	return *claims, nil
}
