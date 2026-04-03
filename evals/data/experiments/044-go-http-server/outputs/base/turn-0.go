package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/time/rate"
)

type URL struct {
	Code      string    `json:"code"`
	Target    string    `json:"target"`
	CreatedAt time.Time `json:"created_at"`
	Clicks    int       `json:"clicks"`
}

type CreateRequest struct {
	Target string `json:"target"`
}

type StatsResponse struct {
	Code   string `json:"code"`
	Clicks int    `json:"clicks"`
}

type Store struct {
	sync.RWMutex
	urls map[string]*URL
}

func (s *Store) Save(u *URL) {
	s.Lock()
	defer s.Unlock()
	s.urls[u.Code] = u
}

func (s *Store) Get(code string) (*URL, bool) {
	s.RLock()
	defer s.RUnlock()
	u, ok := s.urls[code]
	return u, ok
}

func (s *Store) Delete(code string) {
	s.Lock()
	defer s.Unlock()
	delete(s.urls, code)
}

func (s *Store) GetAll() []*URL {
	s.RLock()
	defer s.RUnlock()
	list := make([]*URL, 0, len(s.urls))
	for _, u := range s.urls {
		list = append(list, u)
	}
	return list
}

var store = &Store{urls: make(map[string]*URL)}
var limiter = rate.NewLimiter(1, 3)

func generateCode() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[n.Int64()]
	}
	return string(b)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := time.Now().UnixNano()
		w.Header().Set("X-Request-ID", fmt.Sprintf("%d", id))
		next.ServeHTTP(w, r)
	})
}

func CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	u := &URL{Code: generateCode(), Target: req.Target, CreatedAt: time.Now()}
	store.Save(u)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

func RedirectURL(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")
	u, ok := store.Get(code)
	if !ok {
		http.NotFound(w, r)
		return
	}
	u.Clicks++
	http.Redirect(w, r, u.Target, http.StatusMovedPermanently)
}

func GetStats(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/stats/")
	u, ok := store.Get(code)
	if !ok {
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode(StatsResponse{Code: u.Code, Clicks: u.Clicks})
}

func ListURLs(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(store.GetAll())
}

func DeleteURL(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")
	store.Delete(code)
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", CreateShortURL)
	mux.HandleFunc("GET /stats/", GetStats)
	mux.HandleFunc("GET /urls", ListURLs)
	mux.HandleFunc("DELETE /", DeleteURL)
	mux.HandleFunc("GET /", RedirectURL)

	handler := logging(cors(rateLimit(requestID(mux))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{Addr: ":" + port, Handler: handler}

	go func() {
		log.Printf("Server starting on %s", port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}