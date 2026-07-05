// GET /poll?topic=foo
// client gets latest state immediately or 204
func ShortPolHander(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Query().Get("topic")
		if topic == "" {
			http.Error(w, "Missing topic", http.StatusBadRequest)
			return
		}

		val, ok := store.Get(topic)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(val)

	}
}