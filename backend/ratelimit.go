package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Limitation des essais de mot de passe par adresse IP : au-delà de maxFailures
// échecs dans failureWindow, l'IP attend blockDuration. Tout est en mémoire.
const (
	maxFailures   = 5
	failureWindow = 15 * time.Minute
	blockDuration = 15 * time.Minute
)

type loginEntry struct {
	failures     int
	firstFailure time.Time
	blockedUntil time.Time
}

type loginLimiter struct {
	mu      sync.Mutex
	entries map[string]*loginEntry
}

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{entries: map[string]*loginEntry{}}
}

// Check indique si l'IP est bloquée et pour combien de temps.
func (l *loginLimiter) Check(ip string, now time.Time) (time.Duration, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.purge(now)
	e, ok := l.entries[ip]
	if !ok || now.After(e.blockedUntil) {
		return 0, false
	}
	return e.blockedUntil.Sub(now), true
}

// Fail enregistre un échec et renvoie le nombre d'essais restants avant blocage.
func (l *loginLimiter) Fail(ip string, now time.Time) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.entries[ip]
	if !ok || now.Sub(e.firstFailure) > failureWindow {
		e = &loginEntry{firstFailure: now}
		l.entries[ip] = e
	}
	e.failures++
	if e.failures >= maxFailures {
		e.blockedUntil = now.Add(blockDuration)
		return 0
	}
	return maxFailures - e.failures
}

func (l *loginLimiter) Success(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, ip)
}

// purge oublie les IP dont la fenêtre et le blocage sont passés (appelé sous verrou).
func (l *loginLimiter) purge(now time.Time) {
	for ip, e := range l.entries {
		if now.After(e.blockedUntil) && now.Sub(e.firstFailure) > failureWindow {
			delete(l.entries, ip)
		}
	}
}

// clientIP : derrière nginx + Caddy, l'adresse réelle est la première de X-Forwarded-For.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if first, _, _ := strings.Cut(xff, ","); strings.TrimSpace(first) != "" {
			return strings.TrimSpace(first)
		}
	}
	if rip := strings.TrimSpace(r.Header.Get("X-Real-IP")); rip != "" {
		return rip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func humanWait(d time.Duration) string {
	m := int(d.Minutes())
	if m >= 1 {
		return fmt.Sprintf("%d min", m)
	}
	return fmt.Sprintf("%d s", int(d.Seconds())+1)
}
