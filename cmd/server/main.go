package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func main() {
	// make cache with 10ms TTL and 5 max keys
	cache := expirable.NewLRU[string, string](5, nil, time.Second*60)
	apiMux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./frontend/dist"))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("accept")
		fmt.Println(r.URL.String(), r.Header.Get("accept"))
		if accept != "application/json" {
			if strings.Contains(accept, "text/html") {
				r.URL.Path = "/"
				r.URL.RawPath = "/"
				fileServer.ServeHTTP(w, r)
				return
			}

			fileServer.ServeHTTP(w, r)
			return
		}
		apiMux.ServeHTTP(w, r)
	})
	apiMux.Handle("POST /session", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for {

			id, err := gonanoid.Generate("0123456789", 8)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, err.Error())
				return
			}
			if cache.Contains(id) {
				continue
			}

			cache.Add(id, "")
			w.WriteHeader(http.StatusOK)
			_, err = fmt.Fprint(w, id)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, err.Error())
				return
			}
			return
		}
	}))

	apiMux.Handle("POST /join", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("session")
		if sessionID == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(
				w, "Missing session id")
			return
		}
		gotoURL := r.URL.Query().Get("gotoUrl")
		if gotoURL == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(
				w, "Missing goto URL")
			return
		}

		val, ok := cache.Get(sessionID)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "session not found or expired")
			return
		}

		if val != "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "session already joined")
			return
		}
		cache.Add(sessionID, gotoURL)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, gotoURL)
	}))

	apiMux.Handle("GET /session", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("session")
		if sessionID == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(
				w, "Missing session id")
			return
		}

		gotoURL, ok := cache.Get(sessionID)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "session not found or expired")
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, gotoURL)
	}))

	srv := http.Server{
		Addr:        "0.0.0.0:10000",
		Handler:     handler,
		IdleTimeout: time.Second * 10,
	}
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
