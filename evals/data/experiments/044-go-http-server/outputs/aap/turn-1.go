{
  "protocol": "aap/0.1",
  "id": "url-shortener-service",
  "version": 1,
  "name": "edit",
  "content": [
    {
      "op": "insert_after",
      "target": {
        "type": "id",
        "value": "handlers"
      },
      "content": "func (s *Store) GetTopURLs(w http.ResponseWriter, r *http.Request) {\n\tlimit := 10\n\tif l := r.URL.Query().Get(\"limit\"); l != \"\" {\n\t\tfmt.Sscanf(l, \"%d\", &limit)\n\t}\n\n\ts.mu.RLock()\n\tvar sorted []*URL\n\tfor _, u := range s.urls {\n\t\tsorted = append(sorted, u)\n\t}\n\ts.mu.RUnlock()\n\n\t// Simple sort: most clicked first\n\tfor i := 0; i < len(sorted); i++ {\n\t\tfor j := i + 1; j < len(sorted); j++ {\n\t\t\tif sorted[j].Clicks > sorted[i].Clicks {\n\t\t\t\tsorted[i], sorted[j] = sorted[j], sorted[i]\n\t\t\t}\n\t\t}\n\t}\n\n\tif limit > len(sorted) {\n\t\tlimit = len(sorted)\n\t}\n\tjson.NewEncoder(w).Encode(sorted[:limit])\n}\n"
    },
    {
      "op": "insert_after",
      "target": {
        "type": "id",
        "value": "main-top-routes"
      },
      "content": "\tmux.HandleFunc(\"GET /urls/top\", store.GetTopURLs)"
    }
  ]
}
