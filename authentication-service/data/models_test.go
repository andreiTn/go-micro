package data

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestPasswordMatches(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	u := User{Password: string(hash)}

	ok, err := u.PasswordMatches("secret")
	if err != nil || !ok {
		t.Fatalf("expected match, ok=%v err=%v", ok, err)
	}

	ok, err = u.PasswordMatches("wrong")
	if err != nil || ok {
		t.Fatalf("expected no match, ok=%v err=%v", ok, err)
	}
}
