package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Server struct {
	store   *Store
	auth    *Auth
	limiter *loginLimiter
}

func NewServer(store *Store, auth *Auth) *Server {
	return &Server{store: store, auth: auth, limiter: newLoginLimiter()}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "flavour": "chantilly"})
	})
	mux.HandleFunc("POST /api/login", s.handleLogin)
	mux.Handle("GET /api/me", s.requireAuth(s.handleMe))
	mux.Handle("GET /api/counters", s.requireAuth(s.handleList))
	mux.Handle("POST /api/counters", s.requireAuth(s.handleCreate))
	mux.Handle("GET /api/counters/{id}", s.requireAuth(s.handleGet))
	mux.Handle("PUT /api/counters/{id}", s.requireAuth(s.handleUpdate))
	mux.Handle("DELETE /api/counters/{id}", s.requireAuth(s.handleDelete))
	mux.Handle("POST /api/counters/{id}/pause", s.requireAuth(s.handlePause))
	mux.Handle("POST /api/counters/{id}/resume", s.requireAuth(s.handleResume))
	mux.Handle("POST /api/counters/{id}/reset", s.requireAuth(s.handleReset))
	mux.Handle("POST /api/counters/{id}/archive", s.requireAuth(s.handleArchive))
	mux.Handle("POST /api/counters/{id}/unarchive", s.requireAuth(s.handleUnarchive))
	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "cette route n'existe pas (la crème non plus) 🍮")
	})
	return logRequests(mux)
}

// ---------- helpers ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decode(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	return json.NewDecoder(r.Body).Decode(v)
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, sw.status, time.Since(start).Round(time.Millisecond))
	})
}

// ---------- auth ----------

func (s *Server) requireAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok {
			writeError(w, http.StatusUnauthorized, "il faut le mot de passe pour entrer 🤫")
			return
		}
		if _, valid := s.auth.Verify(strings.TrimSpace(token), time.Now()); !valid {
			writeError(w, http.StatusUnauthorized, "session expirée ou invalide, on se reconnecte ? 🔑")
			return
		}
		next(w, r)
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Password string `json:"password"`
	}
	if err := decode(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "JSON illisible")
		return
	}
	ip := clientIP(r)
	now := time.Now()
	if wait, blocked := s.limiter.Check(ip, now); blocked {
		w.Header().Set("Retry-After", strconv.Itoa(int(wait.Seconds())+1))
		writeError(w, http.StatusTooManyRequests, fmt.Sprintf("trop d'essais, la porte est fermée pendant %s 🫖", humanWait(wait)))
		return
	}
	if !s.auth.CheckPassword(in.Password) {
		left := s.limiter.Fail(ip, now)
		time.Sleep(600 * time.Millisecond) // freine un peu la force brute
		msg := "mauvais mot de passe, la crème a tourné 🙈"
		if left > 0 && left <= 2 {
			msg = fmt.Sprintf("mauvais mot de passe, plus que %d essai(s) avant la pause forcée 🙈", left)
		}
		writeError(w, http.StatusUnauthorized, msg)
		return
	}
	s.limiter.Success(ip)
	token, exp := s.auth.NewToken(now)
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "expiresAt": exp.UTC()})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	exp, _ := s.auth.Verify(strings.TrimSpace(token), time.Now())
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "expiresAt": exp.UTC()})
}

// ---------- counters ----------

type counterInput struct {
	Title   string `json:"title"`
	Phrase  string `json:"phrase"`
	Emoji   string `json:"emoji"`
	Color   string `json:"color"`
	StartAt string `json:"startAt"` // RFC 3339

	GainAmount float64 `json:"gainAmount"`
	GainUnit   string  `json:"gainUnit"`
	GainPeriod string  `json:"gainPeriod"`
	GainLabel  string  `json:"gainLabel"`
	EquivUnit  string  `json:"equivUnit"`
	EquivValue float64 `json:"equivValue"`
}

var allowedPeriods = map[string]bool{"hour": true, "day": true, "week": true, "month": true}

var allowedColors = map[string]bool{
	"cream": true, "pink": true, "lavender": true, "mint": true,
	"sky": true, "peach": true, "lemon": true,
}

// validate nettoie l'entrée et renvoie les champs prêts à être copiés dans un Counter.
func (in counterInput) validate() (Counter, error) {
	var c Counter
	c.Title = strings.TrimSpace(in.Title)
	if c.Title == "" {
		return c, errors.New("il faut un titre, même tout petit 🐣")
	}
	if utf8.RuneCountInString(c.Title) > 80 {
		return c, errors.New("le titre est trop long (80 caractères max)")
	}
	c.Phrase = strings.TrimSpace(in.Phrase)
	if utf8.RuneCountInString(c.Phrase) > 280 {
		return c, errors.New("la phrase est trop longue (280 caractères max)")
	}
	c.Emoji = strings.TrimSpace(in.Emoji)
	if c.Emoji == "" {
		c.Emoji = "🍦"
	}
	if utf8.RuneCountInString(c.Emoji) > 8 {
		return c, errors.New("l'emoji est trop long (un seul suffit 🙂)")
	}
	c.Color = in.Color
	if !allowedColors[c.Color] {
		c.Color = "cream"
	}
	start, err := time.Parse(time.RFC3339, strings.TrimSpace(in.StartAt))
	if err != nil {
		return c, errors.New("la date de départ est invalide (format ISO 8601 attendu)")
	}
	c.StartAt = start.UTC()

	if in.GainAmount < 0 || in.GainAmount > 1e12 {
		return c, errors.New("le gain doit être un nombre positif (et raisonnable 🙂)")
	}
	c.GainAmount = in.GainAmount
	c.GainUnit = strings.TrimSpace(in.GainUnit)
	if utf8.RuneCountInString(c.GainUnit) > 16 {
		return c, errors.New("l'unité du gain est trop longue (16 caractères max)")
	}
	c.GainLabel = strings.TrimSpace(in.GainLabel)
	if utf8.RuneCountInString(c.GainLabel) > 40 {
		return c, errors.New("le libellé du gain est trop long (40 caractères max)")
	}
	c.GainPeriod = in.GainPeriod
	if !allowedPeriods[c.GainPeriod] {
		c.GainPeriod = "day"
	}
	c.EquivUnit = strings.TrimSpace(in.EquivUnit)
	if utf8.RuneCountInString(c.EquivUnit) > 30 {
		return c, errors.New("le nom de l'équivalence est trop long (30 caractères max)")
	}
	if in.EquivValue < 0 || in.EquivValue > 1e12 {
		return c, errors.New("la valeur de l'équivalence doit être un nombre positif")
	}
	c.EquivValue = in.EquivValue
	return c, nil
}

func (s *Server) handleList(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"counters":   s.store.List(),
		"serverTime": time.Now().UTC(),
	})
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	c, err := s.store.Get(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "ce compteur a fondu, il n'existe plus 🫠")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var in counterInput
	if err := decode(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "JSON illisible")
		return
	}
	c, err := in.validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	created, err := s.store.Add(c)
	if err != nil {
		log.Printf("erreur d'écriture : %v", err)
		writeError(w, http.StatusInternalServerError, "impossible d'enregistrer le compteur")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	var in counterInput
	if err := decode(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "JSON illisible")
		return
	}
	clean, err := in.validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := s.store.Update(r.PathValue("id"), func(c *Counter) error {
		c.Title, c.Phrase, c.Emoji, c.Color, c.StartAt = clean.Title, clean.Phrase, clean.Emoji, clean.Color, clean.StartAt
		c.GainAmount, c.GainUnit, c.GainPeriod, c.GainLabel = clean.GainAmount, clean.GainUnit, clean.GainPeriod, clean.GainLabel
		c.EquivUnit, c.EquivValue = clean.EquivUnit, clean.EquivValue
		return nil
	})
	s.respondMutation(w, updated, err)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	err := s.store.Delete(r.PathValue("id"))
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "ce compteur a fondu, il n'existe plus 🫠")
		return
	}
	if err != nil {
		log.Printf("erreur d'écriture : %v", err)
		writeError(w, http.StatusInternalServerError, "impossible de supprimer le compteur")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handlePause(w http.ResponseWriter, r *http.Request) {
	updated, err := s.store.Update(r.PathValue("id"), func(c *Counter) error {
		if c.Paused {
			return nil // déjà en pause : idempotent
		}
		now := time.Now().UTC()
		c.Paused = true
		c.PausedAt = &now
		return nil
	})
	s.respondMutation(w, updated, err)
}

func (s *Server) handleResume(w http.ResponseWriter, r *http.Request) {
	updated, err := s.store.Update(r.PathValue("id"), func(c *Counter) error {
		if !c.Paused {
			return nil
		}
		if c.PausedAt != nil {
			if d := time.Since(*c.PausedAt); d > 0 {
				c.PausedSeconds += int64(d.Seconds())
			}
		}
		c.Paused = false
		c.PausedAt = nil
		return nil
	})
	s.respondMutation(w, updated, err)
}

// handleReset : « j'ai craqué ». La tentative en cours rejoint l'historique et tout repart de maintenant.
func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	updated, err := s.store.Update(r.PathValue("id"), func(c *Counter) error {
		now := time.Now().UTC()
		end := now
		if c.Paused && c.PausedAt != nil {
			end = *c.PausedAt
		}
		secs := int64(end.Sub(c.StartAt).Seconds()) - c.PausedSeconds
		if secs < 0 {
			secs = 0
		}
		c.Attempts = append(c.Attempts, Attempt{StartAt: c.StartAt, EndAt: now, Seconds: secs})
		c.StartAt = now
		c.Paused = false
		c.PausedAt = nil
		c.PausedSeconds = 0
		return nil
	})
	s.respondMutation(w, updated, err)
}

func (s *Server) handleArchive(w http.ResponseWriter, r *http.Request) {
	updated, err := s.store.Update(r.PathValue("id"), func(c *Counter) error {
		if !c.Archived {
			now := time.Now().UTC()
			c.Archived = true
			c.ArchivedAt = &now
		}
		return nil
	})
	s.respondMutation(w, updated, err)
}

func (s *Server) handleUnarchive(w http.ResponseWriter, r *http.Request) {
	updated, err := s.store.Update(r.PathValue("id"), func(c *Counter) error {
		c.Archived = false
		c.ArchivedAt = nil
		return nil
	})
	s.respondMutation(w, updated, err)
}

func (s *Server) respondMutation(w http.ResponseWriter, c Counter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "ce compteur a fondu, il n'existe plus 🫠")
	case err != nil:
		log.Printf("erreur d'écriture : %v", err)
		writeError(w, http.StatusInternalServerError, "impossible d'enregistrer le compteur")
	default:
		writeJSON(w, http.StatusOK, c)
	}
}
