// GET /longpoll?topic=foo
func LongPollHandler(broker *broker.Broker) http.HandlerFunc {
	topic := r.URL.Query().Get("topic")
	timeout := time.Second * 30 // Set a timeout for long polling
	sub := broker.Subscribe(topic)
	defer broker.Unsubscribe(sub)

	select (
	case msg := <-sub.ch:
		// got a message, return it to the client
		w.Header().Set("Content-Type", "application/json")
		w.Write(msg.Payload)
	case <-time.After(timeout):
		// timeout, return 204 No Content
		w.writeHeader(http.StatusNoContent)
	case <-r.Context().Done():
		// client canceled the request
		return
	)