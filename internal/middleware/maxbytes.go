package middleware

import "net/http"

const MaxBytes = 1 << 20

func WithMaxBytes(h http.Handler) http.Handler {
	mxFunc := func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxBytes)
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(mxFunc)
}
