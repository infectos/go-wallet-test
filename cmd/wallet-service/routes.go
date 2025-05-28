package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) getRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/wallet", app.HandleWalletOperation)
		r.Post("/wallets", app.HandleCreateWallet)
		r.Get("/wallets/{wallet_uuid}", app.HandleWalletBalance)
	})

	return r
}
