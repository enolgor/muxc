// Code generated by muxc. DO NOT EDIT.
// versions:
//   muxc v1.0.0
// source: muxc.yaml

package muxc

import (
	"net/http"

	"github.com/enolgor/muxc/examples/basic/controllers"
	"github.com/enolgor/muxc/examples/basic/handlers"
	"github.com/enolgor/muxc/examples/basic/middlewares"
	"github.com/enolgor/muxc/middlewares/logger"
	"log/slog"
)

func chain(f http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}

func stack(mws ...func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		for _, m := range mws {
			f = m(f)
		}
		return f
	}
}

func Middleware(m func(http.Handler) http.Handler) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return m(next).ServeHTTP
	}
}

func ConfigureMux(mux *http.ServeMux, ctrl controllers.Controller) {
	acceptJson := Middleware(middlewares.SetHeader("Accept", "application/json"))
	contentJson := Middleware(middlewares.SetHeader("Content-Type", "application/json"))
	json := stack(contentJson, acceptJson)
	logger := logger.New(slog.Default(), logger.InternalServerError(slog.LevelError), logger.BadRequest(slog.LevelWarn))
	mux.Handle("GET /api/v1/pet", chain(
		handlers.ListPets(ctrl),
		contentJson,
		logger,
		Middleware(middlewares.RequestID),
	))
	mux.Handle("GET /api/v1/pet/{id}", chain(
		handlers.ReadPet(ctrl),
		contentJson,
		logger,
		Middleware(middlewares.RequestID),
	))
	mux.Handle("PUT /api/v1/pet", chain(
		handlers.CreatePet(ctrl),
		json,
		logger,
		Middleware(middlewares.RequestID),
	))
	mux.Handle("POST /api/v1/pet", chain(
		handlers.UpdatePet(ctrl),
		contentJson,
		logger,
		Middleware(middlewares.RequestID),
	))
	mux.Handle("DELETE /api/v1/pet", chain(
		handlers.DeletePet(ctrl),
		logger,
		Middleware(middlewares.RequestID),
	))
	mux.Handle("GET /api/v2/pet", chain(
		handlers.Test(ctrl),
		middlewares.Recover,
		middlewares.InterceptContentSniffer,
		middlewares.InterceptErrorStatus,
		contentJson,
		logger,
	))
}
