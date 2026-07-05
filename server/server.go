func New(broker *broker.Broker, store *store.Store) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/poll", handler.ShortPolHander(store))
	mux.HandleFunc("/longpoll", handler.ShortPolHander(broker))
	mux.HandleFunc("/sse", handler.ShortPolHander(broker))
	mux.HandleFunc("/ws", handler.ShortPolHander(broker))
	mux.HandleFunc("/publish", handler.ShortPolHander(broker, store))

	return &http.Server {
		Addr:		":8000",
		Handler: mux,
		ReadTimeout: 60 * time.Second,
		WrtieTimeout: 0,  // 0 = no timeout on write (SSE/WS streams are unbounded)
		IdleTimeout: 120 * time.Second
	}
}