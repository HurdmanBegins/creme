package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var ErrNotFound = errors.New("compteur introuvable")

// Attempt : une tentative passée (avant un « j'ai craqué »).
type Attempt struct {
	StartAt time.Time `json:"startAt"`
	EndAt   time.Time `json:"endAt"`
	Seconds int64     `json:"seconds"` // durée effective, pauses déduites
}

// Counter : un titre, une phrase, une date de départ… et le temps qui file.
// Le temps écoulé = (pausedAt ou maintenant) - startAt - pausedSeconds.
type Counter struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Phrase        string     `json:"phrase"`
	Emoji         string     `json:"emoji"`
	Color         string     `json:"color"`
	StartAt       time.Time  `json:"startAt"`
	Paused        bool       `json:"paused"`
	PausedAt      *time.Time `json:"pausedAt"`
	PausedSeconds int64      `json:"pausedSeconds"`
	// Gain optionnel : GainAmount par GainPeriod (0 = pas de gain). Ex : 13 € par jour, économisés.
	GainAmount float64 `json:"gainAmount"`
	GainUnit   string  `json:"gainUnit"`
	GainPeriod string  `json:"gainPeriod"` // hour | day | week | month
	GainLabel  string  `json:"gainLabel"`
	// Équivalence rigolote personnalisée : 1 EquivUnit = EquivValue GainUnit (ex. 1 pizza = 12 €).
	EquivUnit  string  `json:"equivUnit"`
	EquivValue float64 `json:"equivValue"`
	// Tentatives précédentes (« j'ai craqué ») et archivage (corbeille réversible).
	Attempts   []Attempt  `json:"attempts"`
	Archived   bool       `json:"archived"`
	ArchivedAt *time.Time `json:"archivedAt"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

// Store : persistance JSON sur disque, protégée par un mutex, écriture atomique.
type Store struct {
	mu       sync.Mutex
	path     string
	counters []*Counter
}

func NewStore(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	s := &Store{path: path, counters: []*Counter{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, s.save()
	}
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(data, &s.counters); err != nil {
		return nil, fmt.Errorf("JSON invalide : %w", err)
	}
	for _, c := range s.counters {
		if c.Attempts == nil {
			c.Attempts = []Attempt{}
		}
	}
	return s, nil
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.counters, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func newID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func (s *Store) List() []Counter {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Counter, 0, len(s.counters))
	for _, c := range s.counters {
		out = append(out, *c)
	}
	return out
}

func (s *Store) Get(id string) (Counter, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.counters {
		if c.ID == id {
			return *c, nil
		}
	}
	return Counter{}, ErrNotFound
}

func (s *Store) Add(c Counter) (Counter, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	c.ID = newID()
	c.CreatedAt = now
	c.UpdatedAt = now
	if c.Attempts == nil {
		c.Attempts = []Attempt{}
	}
	s.counters = append(s.counters, &c)
	if err := s.save(); err != nil {
		s.counters = s.counters[:len(s.counters)-1]
		return Counter{}, err
	}
	return c, nil
}

// Update applique fn au compteur puis persiste ; en cas d'erreur, l'état est restauré.
func (s *Store) Update(id string, fn func(*Counter) error) (Counter, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.counters {
		if c.ID != id {
			continue
		}
		backup := *c
		if err := fn(c); err != nil {
			*c = backup
			return Counter{}, err
		}
		c.UpdatedAt = time.Now().UTC()
		if err := s.save(); err != nil {
			*c = backup
			return Counter{}, err
		}
		return *c, nil
	}
	return Counter{}, ErrNotFound
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, c := range s.counters {
		if c.ID != id {
			continue
		}
		backup := s.counters
		s.counters = append(append([]*Counter{}, s.counters[:i]...), s.counters[i+1:]...)
		if err := s.save(); err != nil {
			s.counters = backup
			return err
		}
		return nil
	}
	return ErrNotFound
}
