package web

import (
	"log"
	"net/http"
	"time"
)

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; script-src 'self'; img-src 'self' data:")
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

type requestActivity struct {
	active map[string]int
}

func newRequestActivity() *requestActivity {
	return &requestActivity{active: make(map[string]int)}
}

func (a *requestActivity) begin(path string) (int, func()) {
	a.active[path]++
	current := a.active[path]
	return current, func() {
		a.active[path]--
		if a.active[path] == 0 {
			delete(a.active, path)
		}
	}
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func requestLog(activity *requestActivity, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		active, done := activity.begin(r.URL.Path)
		defer done()
		sw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(sw, r)
		log.Printf("http method=%s path=%s status=%d duration=%s active=%d", r.Method, r.URL.Path, sw.status, time.Since(start).Round(time.Millisecond), active)
	})
}
