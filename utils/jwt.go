package utils

import (
	"fmt"
	"go-chatter/data"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtClaims struct {
	userID string
	jwt.RegisteredClaims
}

func SetJwt(u *data.User) (string, error) {
	claims := JwtClaims{
		string(u.ID.Hex()),
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return ss, err
}

func GetJwt(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(*JwtClaims)

	if ok {
		return claims.userID, nil
	} else {
		return "", fmt.Errorf("unknown claims type")
	}
}
