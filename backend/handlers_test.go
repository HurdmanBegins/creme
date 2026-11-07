package main

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestValidateCounterInput(t *testing.T) {
	ok := counterInput{Title: "  Sans café ", StartAt: "2026-01-01T08:00:00.000Z", GainAmount: 13, GainUnit: "€", GainPeriod: "fortnight"}
	c, err := ok.validate()
	if err != nil {
		t.Fatalf("entrée valide refusée : %v", err)
	}
	if c.Title != "Sans café" || c.Emoji != "🍦" || c.Color != "cream" || c.GainPeriod != "day" {
		t.Errorf("valeurs par défaut inattendues : %+v", c)
	}
	if !c.StartAt.Equal(time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)) {
		t.Errorf("date mal parsée : %v", c.StartAt)
	}

	bad := []counterInput{
		{Title: "", StartAt: "2026-01-01T08:00:00Z"},
		{Title: "x", StartAt: "hier"},
		{Title: "x", StartAt: "2026-01-01T08:00:00Z", GainAmount: -1},
		{Title: "x", StartAt: "2026-01-01T08:00:00Z", EquivValue: -1},
	}
	for i, in := range bad {
		if _, err := in.validate(); err == nil {
			t.Errorf("cas %d : entrée invalide acceptée : %+v", i, in)
		}
	}
}

func TestAuthTokens(t *testing.T) {
	a := NewAuth("chantilly")
	now := time.Now()
	tok, exp := a.NewToken(now)
	if _, ok := a.Verify(tok, now); !ok {
		t.Fatal("jeton frais refusé")
	}
	if _, ok := a.Verify(tok, exp.Add(time.Second)); ok {
		t.Error("jeton expiré accepté")
	}
	if _, ok := NewAuth("autre").Verify(tok, now); ok {
		t.Error("jeton accepté avec un autre mot de passe")
	}
	if _, ok := a.Verify(tok+"x", now); ok {
		t.Error("jeton altéré accepté")
	}
}

func TestLoginLimiter(t *testing.T) {
	l := newLoginLimiter()
	now := time.Now()
	for i := 1; i < maxFailures; i++ {
		if left := l.Fail("1.2.3.4", now); left != maxFailures-i {
			t.Errorf("essai %d : %d restants, attendu %d", i, left, maxFailures-i)
		}
		if _, blocked := l.Check("1.2.3.4", now); blocked {
			t.Fatalf("bloqué trop tôt (essai %d)", i)
		}
	}
	l.Fail("1.2.3.4", now)
	if wait, blocked := l.Check("1.2.3.4", now); !blocked || wait <= 0 {
		t.Error("pas bloqué après le maximum d'échecs")
	}
	if _, blocked := l.Check("5.6.7.8", now); blocked {
		t.Error("une autre IP est bloquée")
	}
	if _, blocked := l.Check("1.2.3.4", now.Add(blockDuration+time.Second)); blocked {
		t.Error("toujours bloqué après la fin du blocage")
	}
	l.Success("1.2.3.4")
	if _, blocked := l.Check("1.2.3.4", now); blocked {
		t.Error("toujours bloqué après un succès")
	}
}

func TestClientIP(t *testing.T) {
	r := httptest.NewRequest("POST", "/api/login", nil)
	r.RemoteAddr = "172.18.0.5:4321"
	if ip := clientIP(r); ip != "172.18.0.5" {
		t.Errorf("sans en-tête : %q", ip)
	}
	r.Header.Set("X-Forwarded-For", "203.0.113.9, 172.18.0.2")
	if ip := clientIP(r); ip != "203.0.113.9" {
		t.Errorf("avec X-Forwarded-For : %q", ip)
	}
}
