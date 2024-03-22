package middlewares

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

var Logger func(http.Handler) http.Handler = middleware.Logger
var RequestID func(http.Handler) http.Handler = middleware.RequestID
var SetHeader = middleware.SetHeader
