package whatsapp

import "testing"

func TestVerifySignature(t *testing.T) {
	body := []byte(`{"event":"message.received"}`)
	secret := "0123456789abcdef-secret"
	sig := Sign(body, secret)
	tests := []struct {
		name   string
		body   []byte
		header string
		secret string
		want   bool
	}{
		{"valid", body, sig, secret, true},
		{"tampered body", []byte(`{"event":"x"}`), sig, secret, false},
		{"wrong secret", body, sig, "another-secret-1234", false},
		{"no prefix", body, sig[len("sha256="):], secret, false},
		{"empty header", body, "", secret, false},
		{"empty secret", body, sig, "", false},
		{"garbage", body, "sha256=zz", secret, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifySignature(tt.body, tt.header, tt.secret); got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
