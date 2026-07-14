package utils

import "testing"

func TestHashPasswordProducesDifferentHashesEachTime(t *testing.T) {
	hash1, err := HashPassword("s3cret-pass")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	hash2, err := HashPassword("s3cret-pass")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if hash1 == hash2 {
		t.Fatal("HashPassword returned identical hashes for the same input; bcrypt should salt each call")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	hash, err := HashPassword("s3cret-pass")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if !CheckPasswordHash("s3cret-pass", hash) {
		t.Error("CheckPasswordHash returned false for the correct password")
	}
	if CheckPasswordHash("wrong-pass", hash) {
		t.Error("CheckPasswordHash returned true for an incorrect password")
	}
}
