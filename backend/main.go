// Crême 2.0 — backend minimaliste : un mot de passe, un fichier JSON, des compteurs.
package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	password := os.Getenv("APP_PASSWORD")
	if password == "" {
		log.Fatal("APP_PASSWORD est vide : définis-le dans le .env")
	}
	if password == "change-moi-vraiment" || password == "chantilly" {
		log.Fatal("APP_PASSWORD a encore la valeur d'exemple : choisis un vrai mot de passe dans le .env")
	}
	if len(password) < 8 {
		log.Fatal("APP_PASSWORD est trop court (8 caractères minimum)")
	}
	dataFile := env("DATA_FILE", "/data/counters.json")
	addr := ":" + env("BACKEND_PORT", "8080")

	store, err := NewStore(dataFile)
	if err != nil {
		log.Fatalf("impossible d'ouvrir %s : %v", dataFile, err)
	}

	srv := NewServer(store, NewAuth(password))

	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	log.Printf("🍦 Crême 2.0 à l'écoute sur %s (données : %s)", addr, dataFile)
	log.Fatal(httpSrv.ListenAndServe())
}
