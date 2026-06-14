package main

import "net/http"

func (app *application) handle() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/healthcheck", app.healthcheck)

	return mux
}
