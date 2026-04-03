<aap:target id="url-shortener-service">package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

<aap:target id="types">
type URL struct {
	Code      string    `json:"code"`
	LongURL   string    `json:"long_url"`
	CreatedAt time.Time `json:"created_at"`
	Clicks    int       `json:"clicks"`
}

type CreateRequest struct {
	LongURL string `json:"long_url"`
}

type StatsResponse struct {
	Code    string `json:"code"`
	LongURL string `json:"long_url"`
	Clicks  int    `json:"clicks"`
}
</aap:target>

<aap:target id="store">
type Store struct {
	mu   sync.RWMutex
	urls map[string]*URL
}

func NewStore() *Store {
	return &Store{urls: make(map[string]*URL)}
}
</aap:target>

<aap:target id="middleware">
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "requestID", uuid.New().String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

var limiter = rate.NewLimiter(1, 3)

func rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
</aap:target>

<aap:target id="handlers">
func (s *Store) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	code := uuid.New().String()[:8]
	s.mu.Lock()
	s.urls[code] = &URL{Code: code, LongURL: req.LongURL, CreatedAt: time.Now()}
	s.mu.Unlock()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s.urls[code])
}

func (s *Store) RedirectURL(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]
	s.mu.Lock()
	entry, ok := s.urls[code]
	if ok {
		entry.Clicks++
	}
	s.mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, entry.LongURL, http.StatusMovedPermanently)
}

func (s *Store) GetStats(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[7:]
	s.mu.RLock()
	entry, ok := s.urls[code]
	s.mu.RUnlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode(StatsResponse{Code: entry.Code, LongURL: entry.LongURL, Clicks: entry.Clicks})
}
</aap:target>

<aap:target id="main">
func main() {
	store := NewStore()
	mux := http.NewServeMux()

	mux.HandleFunc("POST /", store.CreateShortURL)
	mux.HandleFunc("GET /stats/", store.GetStats)
	mux.HandleFunc("GET /", store.RedirectURL)

	handler := logging(requestID(rateLimit(mux)))

	srv := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
</aap:target>
</aap:target>