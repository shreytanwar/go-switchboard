// GET /sse?topic=foo

func SSEHandler(broker * broker.Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r * http.Request) {
		topic := r.URL.Query().Get("topic")


        // SSE required headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection". "keep-alive")

		//TODO: Flusher + syntax explain
		fusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		sub := broker.Subscribe(topic)
		defer broker.Unsubscribe(sub)

		for {
			select {
			case msg:= <= sub.ch;
			//TODO
			fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
			flusher.Flush()

			case <- r.Context().Done():
				return
			}
		}
	}
}

// Client auto-reconnects natively via EventSource in browsers. Only standard library needed — SSE is just HTTP with specific headers + \n\n delimited text frames.