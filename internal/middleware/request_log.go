package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	code int
	n    int
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.code = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	if rr.code == 0 {
		rr.code = http.StatusOK
	}
	n, err := rr.ResponseWriter.Write(b)
	rr.n += n
	return n, err
}

// RequestLog logs each HTTP request: method, URI, status, response size, duration, and remote address.
func RequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rr := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(rr, r)
		code := rr.code
		if code == 0 {
			code = http.StatusOK
		}
		log.Printf("%s %s %d %dB %s %s",
			r.Method, r.URL.RequestURI(), code, rr.n, time.Since(start), r.RemoteAddr)
	})
}
