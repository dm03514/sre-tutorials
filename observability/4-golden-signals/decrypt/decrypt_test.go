package decrypt

import (
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestBcrypter_IsMatch_NoMatch(t *testing.T) {
	b := Bcrypter{
		HashedPassword: "$2b$10$//DXiVVE59p7G5k/4Klx/ezF7BI42QZKmoOD0NDvUuqxRE5bFFBLy",
		Password:       "nomatch",
	}
	if b.IsMatch() {
		t.Errorf("expected no match, got match")
	}
}

func TestBcrypter_IsMatch_True(t *testing.T) {
	password := []byte("hasmatch")
	hash, err := bcrypt.GenerateFromPassword(password, 10)
	if err != nil {
		t.Error(err)
	}
	b := Bcrypter{
		HashedPassword: string(hash),
		Password:       string(password),
	}
	if !b.IsMatch() {
		t.Errorf("expected match, got no match")
	}
}
