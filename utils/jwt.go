package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(email string, userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":  email,
		"userID": userID,
		"exp":    time.Now().Add(time.Hour * 2).Unix(), // Token expires in 24 hours
	})
	return token.SignedString([]byte(GetEnv("SECRET_KEY", "sdfjs98sdf9sdfk")))
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(GetEnv("SECRET_KEY", "sdfjs98sdf9sdfk")), nil
	})
}

// UserIDFromToken reads the userID claim set by GenerateJWT out of a token
// already verified by ValidateJWT. ok is false if the token has no usable
// userID claim (JSON numbers decode as float64, hence the type assertion).
func UserIDFromToken(token *jwt.Token) (int64, bool) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, false
	}
	userID, ok := claims["userID"].(float64)
	if !ok {
		return 0, false
	}
	return int64(userID), true
}
