package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateJWTRoundTrip(t *testing.T) {
	tokenString, err := GenerateJWT("alice@gmail.com", 42)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	token, err := ValidateJWT(tokenString)
	if err != nil {
		t.Fatalf("ValidateJWT: %v", err)
	}
	if !token.Valid {
		t.Fatal("ValidateJWT returned a token marked invalid without an error")
	}

	userID, ok := UserIDFromToken(token)
	if !ok {
		t.Fatal("UserIDFromToken: no userID claim found")
	}
	if userID != 42 {
		t.Errorf("UserIDFromToken = %d, want 42", userID)
	}
}

func TestValidateJWTRejectsGarbage(t *testing.T) {
	if _, err := ValidateJWT("not.a.jwt"); err == nil {
		t.Fatal("ValidateJWT accepted a malformed token")
	}
}

func TestValidateJWTRejectsTamperedSignature(t *testing.T) {
	tokenString, err := GenerateJWT("alice@gmail.com", 42)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		t.Fatalf("token has %d dot-separated parts, want 3", len(parts))
	}

	// Flip the first character of the signature, not the last: base64url's
	// final character can carry unused padding bits, so changing only it
	// sometimes decodes to the same bytes and the "tampered" signature
	// still verifies.
	sig := []byte(parts[2])
	if sig[0] == 'x' {
		sig[0] = 'y'
	} else {
		sig[0] = 'x'
	}
	parts[2] = string(sig)
	tampered := strings.Join(parts, ".")

	if tampered == tokenString {
		t.Fatal("test setup did not actually change the token")
	}

	if _, err := ValidateJWT(tampered); err == nil {
		t.Fatal("ValidateJWT accepted a token with a tampered signature")
	}
}

func TestValidateJWTRejectsExpiredToken(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":  "alice@gmail.com",
		"userID": 42,
		"exp":    time.Now().Add(-time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(GetEnv("SECRET_KEY", "sdfjs98sdf9sdfk")))
	if err != nil {
		t.Fatalf("SignedString: %v", err)
	}

	if _, err := ValidateJWT(tokenString); err == nil {
		t.Fatal("ValidateJWT accepted an expired token")
	}
}

// TestValidateJWTRejectsNoneAlgorithm guards against the classic JWT "alg
// confusion" attack, where a token claims alg "none" (no signature check
// at all) to bypass verification. ValidateJWT's keyfunc must reject any
// non-HMAC algorithm before it ever gets to comparing signatures.
func TestValidateJWTRejectsNoneAlgorithm(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"email":  "alice@gmail.com",
		"userID": 42,
		"exp":    time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("SignedString: %v", err)
	}

	if _, err := ValidateJWT(tokenString); err == nil {
		t.Fatal("ValidateJWT accepted a token signed with the \"none\" algorithm")
	}
}
