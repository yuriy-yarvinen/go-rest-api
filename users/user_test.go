package users

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUserMarshalJSONNeverIncludesPassword(t *testing.T) {
	u := User{ID: 1, Email: "alice@gmail.com", Password: "either-a-hash-or-plaintext"}

	out, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if strings.Contains(string(out), "password") || strings.Contains(string(out), u.Password) {
		t.Errorf("Marshal(User) = %s, must never include the password field or value", out)
	}
}

func TestUserUnmarshalStillReadsPassword(t *testing.T) {
	var u User
	if err := json.Unmarshal([]byte(`{"email":"bob@gmail.com","password":"plaintext"}`), &u); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if u.Password != "plaintext" {
		t.Errorf("Password = %q, want %q (MarshalJSON must not affect input binding)", u.Password, "plaintext")
	}
}
