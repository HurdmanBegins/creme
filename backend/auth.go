package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

// sessionTTL : durée de vie d'un jeton de session.
const sessionTTL = 30 * 24 * time.Hour

// Auth gère l'unique mot de passe et des jetons signés (HMAC) sans état :
// ils survivent aux redémarrages et deviennent invalides si le mot de passe change.
type Auth struct {
	password []byte
	key      []byte
}

func NewAuth(password string) *Auth {
	sum := sha256.Sum256([]byte("creme-2.0-session|" + password))
	return &Auth{password: []byte(password), key: sum[:]}
}

func (a *Auth) CheckPassword(p string) bool {
	return subtle.ConstantTimeCompare([]byte(p), a.password) == 1
}

func (a *Auth) sign(payload string) string {
	m := hmac.New(sha256.New, a.key)
	m.Write([]byte(payload))
	return hex.EncodeToString(m.Sum(nil))
}

// NewToken renvoie "<base64(expiration unix)>.<hmac>" et sa date d'expiration.
func (a *Auth) NewToken(now time.Time) (string, time.Time) {
	exp := now.Add(sessionTTL)
	payload := base64.RawURLEncoding.EncodeToString([]byte(strconv.FormatInt(exp.Unix(), 10)))
	return payload + "." + a.sign(payload), exp
}

// Verify valide la signature puis l'expiration du jeton.
func (a *Auth) Verify(token string, now time.Time) (time.Time, bool) {
	payload, sig, ok := strings.Cut(token, ".")
	if !ok {
		return time.Time{}, false
	}
	if subtle.ConstantTimeCompare([]byte(sig), []byte(a.sign(payload))) != 1 {
		return time.Time{}, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return time.Time{}, false
	}
	expUnix, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	exp := time.Unix(expUnix, 0)
	return exp, now.Before(exp)
}
