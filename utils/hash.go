package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	bcryptPasswordInBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bcryptPasswordInBytes), nil
}

// CheckPasswordHash reports whether password matches the bcrypt hash.
// bcrypt salts each hash randomly, so hashes can never be compared by
// equality — they must be verified with bcrypt's own comparison.
func CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
