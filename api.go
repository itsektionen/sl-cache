package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func API() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	// r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.AllowAll().Handler)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("SL Cache. Checkout /sites or /ittv"))
	})

	r.Get("/sites", getSites)
	r.Get("/ittv", getITTV)

	http.ListenAndServe(":3001", r)
}

func getSites(w http.ResponseWriter, r *http.Request) {
	ApiCache.Mutex.RLock()
	defer ApiCache.Mutex.RUnlock()
	if !ApiCache.IsValid {
		w.WriteHeader(500)
		w.Write([]byte("Cache is invalid :("))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(ApiCache.Sites))
}
func getITTV(w http.ResponseWriter, r *http.Request) {
	ApiCache.Mutex.RLock()
	defer ApiCache.Mutex.RUnlock()
	if !ApiCache.IsValid {
		w.WriteHeader(500)
		w.Write([]byte("Cache is invalid :("))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(ApiCache.ITTV))
}
