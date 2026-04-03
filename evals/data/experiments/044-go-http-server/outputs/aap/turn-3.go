package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
)

<aap:target id="rate-limiter">
type ipLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

func newIPLimiter() *ipLimiter {
	return &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

func (i *ipLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		<aap:target id="rate-limit-config">
		limiter = rate.NewLimiter(5, 10)
		</aap:target>
		i.limiters[ip] = limiter
	}
	return limiter
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	limiter := newIPLimiter()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !limiter.getLimiter(ip).Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
</aap:target>

<aap:target id="url-struct">
type URL struct {
	Code      string    `json:"code"`
	Target    string    `json:"target"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}
</aap:target>

<aap:target id="url-store">
var (
	urls = make(map[string]*URL)
	mu   sync.RWMutex
)
</aap:target>

<aap:target id="handlers">
func ExpireURLHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]

	mu.Lock()
	defer mu.Unlock()

	url, exists := urls[code]
	if !exists {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	url.ExpiresAt = time.Now()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "URL %s expired", code)
}
</aap:target>

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/{code}/expire", ExpireURLHandler).Methods("PATCH")
}
