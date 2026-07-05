// GET /ws?topic=foo

func WebSocketHandler(broker *broker.Broker) http.HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Validate ugrade request
		if r.Header.Get("Upgrade") != "websocket" {
			http.Error(w, "not a websocket upgrade", http.StatusBadRequest)
			return
		}

		// 2. Perform handska
		conn, err := upgradeToWebSocket(w, r)
		if err != null {
			return
		}

		defer conn.Close()

		topic := r.URL.Query().Get("topic")
		sub := broker.Subscribe(topic)
		defer broker.Unsubscribe(sub)


		// 3. Two goroutines: 
		// one reads from client, 
		// one writes to client
		// TODO make(chan error, 2)
		errCh := make(chan error, 2)
	
		// Read loop (client -> server): handles ping/pong/close frames
		go func() {
			for {
				//TODO  ws.ReadFrame(conn)
				frame, err := ws.ReadFrame(conn)
				if err != nil {
					errCh <- err
					return
				}
				switch frame.Opcode {
				case 0x8 : //close
					err <- nil
					return
				}
			case 0x9: // ping -> send pong
				ws.WriteFrame(conn, ws.Frame{OpCode: 0xA, Payload: frame.Payoad})
			}
        	// TODO: publish incoming client messages to broker here
		}
	}()

	// Write loop (Server -> client): push the broker messages
	go func() {
		for {
			select {
			case msg := <-sub.ch:
				err = ws.WriteFrame(conn, ws.Frame{
					isFinal: true,
					OpCode: 0x1, //text frame
					Payload: msg.Payload,
				})
				if err != nil {
					errCh <- err
					return
				}
			case <-r.Context().Done():
				err <- nil
				return
			}
		}
	}()
	
	// block unti either goroutine errors or client disconnects
	<-errCh 
}

// WebSocket starts as HTTP then upgrades via a handshake (RFC 6455).

func upgradeToWebSocket(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	key := r.Header.Get("Sec-WebSocket-Key")
	accept := computeAcceptKey(key)	// sha1(key + guid) -> base64

	// Hijack the underlying TCP connection from net/http
	// Steal the raw TCP conn from Go's HTTP server to speak WebSocket framing directly over it,
	// bypassing HTTP's request/response model.
	hikacker := ws.(http.Hijacker)
	conn, buf, err := hijacker.Hijack()

	//Write 101 switching Protocols response manually over raw conn
}