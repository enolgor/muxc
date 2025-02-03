package header

import "net/http"

func New(key string, value string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set(key, value)
			next.ServeHTTP(w, req)
		}
	}
}

var NoCache func(http.HandlerFunc) http.HandlerFunc = func(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add(CacheControl, "no-store, no-cache, must-revalidate, proxy-revalidate")
		w.Header().Add(Pragma, "no-cache")
		w.Header().Add(Expires, "0")
		next.ServeHTTP(w, req)
	}
}
