package auth

import (
	"testing"
	"time"
)

// RFC 6238 appendix B vectors (SHA1, 8 digits) truncated to 6 digits.
func TestTOTPRFCVectors(t *testing.T) {
	key := []byte("12345678901234567890")
	cases := []struct {
		unix int64
		want string
	}{
		{59, "287082"},
		{1111111109, "081804"},
		{1111111111, "050471"},
		{1234567890, "005924"},
		{2000000000, "279037"},
	}
	for _, c := range cases {
		if got := totpAt(key, c.unix/30); got != c.want {
			t.Errorf("t=%d: got %s want %s", c.unix, got, c.want)
		}
	}
}

func TestVerifyTOTP(t *testing.T) {
	secret, err := NewTOTPSecret()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_700_000_000, 0)
	code, _ := TOTPCode(secret, now)
	step, ok := VerifyTOTP(secret, code, now)
	if !ok || step != now.Unix()/30 {
		t.Fatalf("current code rejected")
	}
	if _, ok := VerifyTOTP(secret, code, now.Add(30*time.Second)); !ok {
		t.Error("previous step should be accepted (clock skew)")
	}
	if _, ok := VerifyTOTP(secret, code, now.Add(90*time.Second)); ok {
		t.Error("code three steps old must be rejected")
	}
	if _, ok := VerifyTOTP(secret, "12345", now); ok {
		t.Error("short code accepted")
	}
	spaced := code[:3] + " " + code[3:]
	if _, ok := VerifyTOTP(secret, spaced, now); !ok {
		t.Error("code with a space should be accepted")
	}
}

func TestSealer(t *testing.T) {
	s, err := NewSealer([]byte("k"))
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := s.Seal([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Open(sealed)
	if err != nil || string(got) != "secret" {
		t.Fatalf("roundtrip: %q %v", got, err)
	}
	sealed[len(sealed)-1] ^= 1
	if _, err := s.Open(sealed); err == nil {
		t.Error("tampered ciphertext must fail")
	}
	other, _ := NewSealer([]byte("other"))
	sealed2, _ := s.Seal([]byte("x"))
	if _, err := other.Open(sealed2); err == nil {
		t.Error("wrong key must fail")
	}
}
