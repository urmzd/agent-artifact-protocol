package main

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type ipLimiter struct {
	sync.Mutex
	ips map[string]*rate.Limiter
}

var ipStore = &ipLimiter{
	ips: make(map[string]*rate.Limiter),
}

func getLimiter(ip string) *rate.Limiter {
	ipStore.Lock()
	defer ipStore.Unlock()

	if limiter, exists := ipStore.ips[ip]; exists {
		return limiter
	}

	// 5 requests per second, burst of 10
	limiter := rate.NewLimiter(5, 10)
	ipStore.ips[ip] = limiter
	return limiter
}

func rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}

		limiter := getLimiter(host)
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}