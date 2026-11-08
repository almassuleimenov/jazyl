// Package auth provides password hashing, JWT access tokens and opaque random tokens.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// ErrInvalidToken is returned for any token that fails verification.
var ErrInvalidToken = errors.New("invalid token")

// HashPassword hashes a password with bcrypt.
func HashPassword(pw string) (string, error) {
	if len(pw) < 10 {
		return "", errors.New("password must be at least 10 characters")
	}
	if len(pw) > 72 {
		return "", errors.New("password must be at most 72 bytes")
	}
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(h), nil
}

// dummyHash is compared against when a user does not exist, to keep login timing uniform.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("timing-equaliser-password"), bcryptCost)

// CheckPassword reports whether pw matches hash. Pass an empty hash for unknown users.
func CheckPassword(hash, pw string) bool {
	if hash == "" {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(pw)) // constant-time padding only
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// Claims carried by the access token.
type Claims struct {
	Role     string `json:"role"`
	ClinicID string `json:"cid,omitempty"`
	Purpose  string `json:"pur,omitempty"` // "" = access, "ws" = websocket ticket
	jwt.RegisteredClaims
}

// Principal is the authenticated caller.
type Principal struct {
	UserID   uuid.UUID
	Role     string
	ClinicID *uuid.UUID
}

// Tokens issues and verifies HS256 JWTs.
type Tokens struct {
	secret []byte
	issuer string
}

// NewTokens creates a token issuer.
func NewTokens(secret []byte) *Tokens { return &Tokens{secret: secret, issuer: "jazyl"} }

// Issue creates a signed token for p valid for ttl.
func (t *Tokens) Issue(p Principal, purpose string, ttl time.Duration) (string, error) {
	now := time.Now()
	c := Claims{
		Role:    p.Role,
		Purpose: purpose,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    t.issuer,
			Subject:   p.UserID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.NewString(),
		},
	}
	if p.ClinicID != nil {
		c.ClinicID = p.ClinicID.String()
	}
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(t.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return s, nil
}

// Verify parses a token and checks it was issued for purpose.
func (t *Tokens) Verify(token, purpose string) (Principal, error) {
	var c Claims
	_, err := jwt.ParseWithClaims(token, &c, func(*jwt.Token) (any, error) { return t.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(t.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil || c.Purpose != purpose {
		return Principal{}, ErrInvalidToken
	}
	uid, err := uuid.Parse(c.Subject)
	if err != nil {
		return Principal{}, ErrInvalidToken
	}
	p := Principal{UserID: uid, Role: c.Role}
	if c.ClinicID != "" {
		cid, err := uuid.Parse(c.ClinicID)
		if err != nil {
			return Principal{}, ErrInvalidToken
		}
		p.ClinicID = &cid
	}
	return p, nil
}

// RandomToken returns a URL-safe random token with prefix, plus its SHA-256 hash for storage.
func RandomToken(prefix string, nBytes int) (token string, hash []byte, err error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", nil, fmt.Errorf("random: %w", err)
	}
	token = prefix + base64.RawURLEncoding.EncodeToString(b)
	return token, HashToken(token), nil
}

// HashToken hashes an opaque token for lookup. Tokens are high-entropy, so SHA-256 is enough.
func HashToken(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}
