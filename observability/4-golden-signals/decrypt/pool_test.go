package decrypt

import "testing"

func TestPool_IsMatch_FalseNoMatch(t *testing.T) {
	pool := NewPool(1)
	b := Bcrypter{
		HashedPassword: "$2b$10$//DXiVVE59p7G5k/4Klx/ezF7BI42QZKmoOD0NDvUuqxRE5bFFBLy",
		Password:       "nomatch",
	}
	// more explicit error assertion?
	if pool.IsMatch(b) != false {
		t.Errorf("expected no match, received match")
	}
}
