// Package phone normalises phone numbers to digits-only international form (Kazakhstan-first).
package phone

import (
	"errors"
	"strings"
)

// ErrInvalid is returned for numbers that cannot be normalised.
var ErrInvalid = errors.New("invalid phone number")

// Normalize converts "+7 (701) 123-45-67", "87011234567", "7011234567" into "77011234567".
func Normalize(raw string) (string, error) {
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	d := b.String()
	switch {
	case len(d) == 11 && d[0] == '8':
		d = "7" + d[1:] // local KZ/RU trunk prefix
	case len(d) == 10 && (d[0] == '7' || d[0] == '6'):
		d = "7" + d // missing country code: 701..., 747..., 6xx
	}
	if len(d) < 10 || len(d) > 15 {
		return "", ErrInvalid
	}
	return d, nil
}

// FromChatID extracts the phone from a WhatsApp chat id ("77011234567@c.us").
// Group chats and LID addresses return ok=false.
func FromChatID(chatID string) (string, bool) {
	local, host, found := strings.Cut(chatID, "@")
	if found && host != "c.us" && host != "s.whatsapp.net" {
		return "", false
	}
	local, _, _ = strings.Cut(local, ":") // device suffix "7701...:12"
	p, err := Normalize(local)
	if err != nil {
		return "", false
	}
	return p, true
}

// ChatID returns the WhatsApp chat id for a normalised phone.
func ChatID(p string) string { return p + "@c.us" }

// Mask hides the middle of a number for logs: 7701***4567.
func Mask(p string) string {
	if len(p) < 8 {
		return "***"
	}
	return p[:4] + "***" + p[len(p)-4:]
}
