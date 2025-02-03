package recoverer

import "net/http"

func New(handler func(panicked any, w http.ResponseWriter, req *http.Request)) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					handler(r, w, req)
				}
			}()
			next(w, req)
		}
	}
}
