func GetTopURLs(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	all := store.GetAll()
	sort.Slice(all, func(i, j int) bool {
		return all[i].Clicks > all[j].Clicks
	})

	if limit > len(all) {
		limit = len(all)
	}

	json.NewEncoder(w).Encode(all[:limit])
}

// In main:
// mux.HandleFunc("GET /urls/top", GetTopURLs)

// Ensure imports include: "sort", "strconv"