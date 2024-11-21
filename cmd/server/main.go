package main

import (
	"net/http"

	"github.com/diegopontes87/multithreading-challenge/infra/webserver/handlers"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/{cep}", handlers.GetCEPInfo)
	http.ListenAndServe(":8080", r)
}
