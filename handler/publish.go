// POST /publish
// body: {"topic": "foo", "payload": "hello"}

func PublishHandler(broker *broker.Broker store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Topic string `json:"topic"`
			Payload []byte `json:"payload"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		// write to both: store (short poll) and broker (for streaming transports)
		store.Set(req.Topic, req.Payload)
		broker.Publish(req.Topic, req.Payload)

		w.writeHeader(http.StatusAccepted)
	}
}