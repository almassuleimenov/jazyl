package phone

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		in, want string
		wantErr  bool
	}{
		{"+7 (701) 123-45-67", "77011234567", false},
		{"87011234567", "77011234567", false},
		{"7011234567", "77011234567", false},
		{"77011234567", "77011234567", false},
		{"8 747 123 45 67", "77471234567", false},
		{"998901234567", "998901234567", false},
		{"12345", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := Normalize(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFromChatID(t *testing.T) {
	tests := []struct {
		in   string
		want string
		ok   bool
	}{
		{"77011234567@c.us", "77011234567", true},
		{"77011234567:12@s.whatsapp.net", "77011234567", true},
		{"120363021234567890@g.us", "", false},
		{"123456789012345@lid", "", false},
	}
	for _, tt := range tests {
		got, ok := FromChatID(tt.in)
		if ok != tt.ok || got != tt.want {
			t.Errorf("FromChatID(%q) = %q,%v want %q,%v", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}

func FuzzNormalize(f *testing.F) {
	f.Add("+7 701 123 45 67")
	f.Add("8(747)1234567")
	f.Fuzz(func(t *testing.T, s string) {
		p, err := Normalize(s)
		if err == nil && (len(p) < 10 || len(p) > 15) {
			t.Fatalf("invalid length %q", p)
		}
	})
}
