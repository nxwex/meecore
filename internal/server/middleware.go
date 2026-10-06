package server

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(body []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}

	return w.ResponseWriter.Write(body)
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("http.Hijacker is not supported")
	}

	return hijacker.Hijack()
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Printf(
			"[IN]  %s %s remote_ip:[%s]",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
		)

		writer := &responseWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(writer, r)

		log.Printf(
			"[OUT] %s %s [%d] [%s]",
			r.Method,
			r.URL.Path,
			writer.statusCode,
			time.Since(start).Round(time.Millisecond),
		)
	})
}
