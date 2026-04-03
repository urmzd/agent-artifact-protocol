type URL struct {
	Code      string     `json:"code"`
	Target    string     `json:"target"`
	CreatedAt time.Time  `json:"created_at"`
	Clicks    int        `json:"clicks"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type ExpireRequest struct {
	ExpiresAt time.Time `json:"expires_at"`
}

func UpdateExpiration(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")
	code = strings.TrimSuffix(code, "/expire")

	u, ok := store.Get(code)
	if !ok {
		http.NotFound(w, r)
		return
	}

	var req ExpireRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u.ExpiresAt = &req.ExpiresAt
	store.Save(u)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

// Inside RedirectURL, add expiration check:
func RedirectURL(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")
	u, ok := store.Get(code)
	if !ok {
		http.NotFound(w, r)
		return
	}

	if u.ExpiresAt != nil && time.Now().After(*u.ExpiresAt) {
		http.Error(w, "URL expired", http.StatusGone)
		return
	}

	u.Clicks++
	http.Redirect(w, r, u.Target, http.StatusMovedPermanently)
}

// In main, register:
// mux.HandleFunc("PATCH /", UpdateExpiration)