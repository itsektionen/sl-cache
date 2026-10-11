package main

import (
	"encoding/json"
	"net/http"
	"strconv"

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
		w.Write([]byte("SL Cache. Checkout /sites or /site/{id}"))
	})

	r.Get("/site/{id}", getSite)
	r.Get("/sites", getSites)

	http.ListenAndServe(":3001", r)
}

func getSite(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Please provide a site id! You can find a list of sites here: https://transport.integration.sl.se/v1/sites?expand=true", 401)
		return
	}
	idNum, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Please provide a site id! You can find a list of sites here: https://transport.integration.sl.se/v1/sites?expand=true", 401)
		return
	}
	cachedData, exists := GetSite(int(idNum))
	if !exists {
		http.Error(w, "Site not found, add it to the config. If the server recently restarted then it will take a bit for all sites to be fetched.", 400)
		return
	}
	d, err := json.Marshal(cachedData.Departures)
	if err != nil {
		println("ERROR! Could not convert Departures to JSON! " + err.Error())
		http.Error(w, "Internal server error", 500)
		return
	}
	w.Write([]byte(d))
}

func getSites(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(GetSites()))
}
