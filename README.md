# go-switchboard
Different kind of connection for go server (standard libraries)

| Concern | How handled |
|---|---|
| Broker fan-out under concurrent publishes | `sync.RWMutex` — concurrent reads ok, exclusive on subscribe/unsubscribe |
| Slow consumer in broker | Buffered channels + non-blocking send — drop message rather than stall |
| WebSocket concurrent read+write | Two goroutines with a single error channel — clean exit on either side failing |
| SSE client disconnect | `r.Context().Done()` channel — goroutine exits and unsubscribes cleanly |
| Long poll timeout | `time.After` in select — releases goroutine and connection predictably |
 


 # switchboard

A Go server demonstrating every major real-time transport pattern — short poll,
long poll, SSE, WebSocket, and gRPC — all sharing one in-memory pub/sub broker.

## Transports

| Transport      | Endpoint              | Direction      | Notes                              |
|----------------|-----------------------|----------------|------------------------------------|
| Short poll     | `GET /poll`           | server→client  | Returns latest value immediately   |
| Long poll      | `GET /longpoll`       | server→client  | Held open 30s or until message     |
| SSE            | `GET /sse`            | server→client  | Persistent stream, auto-reconnect  |
| WebSocket      | `GET /ws`             | bidirectional  | Full RFC 6455, no external lib     |
| Publish        | `POST /publish`       | client→server  | Fans out to all transports         |
| gRPC Publish   | `Publish` RPC         | unary          | Single publish + ack               |
| gRPC Subscribe | `Subscribe` RPC       | server-stream  | Like SSE but over HTTP/2           |
| gRPC Chat      | `Chat` RPC            | bidirectional  | Like WebSocket but over HTTP/2     |

## Run (HTTP only, stdlib)

```bash
go run .
# HTTP on :8080
```

## Run (with gRPC)

1. Install protoc plugins:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

2. Generate protobuf code:
```bash
protoc --go_out=. --go_opt=module=github.com/you/switchboard \
       --go-grpc_out=. --go-grpc_opt=module=github.com/you/switchboard \
       grpc/proto/switchboard.proto
```

3. Add dependencies:
```bash
go get google.golang.org/grpc google.golang.org/protobuf
```

4. Uncomment the gRPC block in `main.go`, then:
```bash
go run -tags grpc .
# HTTP on :8080, gRPC on :9090
```

## Quick test

```bash
# Publish a message
curl -X POST http://localhost:8080/publish \
  -H "Content-Type: application/json" \
  -d '{"topic":"news","payload":{"headline":"hello"}}'

# Short poll (fire and forget)
curl http://localhost:8080/poll?topic=news

# Long poll (blocks until message arrives)
curl http://localhost:8080/longpoll?topic=news

# SSE stream (open a persistent connection)
curl -N http://localhost:8080/sse?topic=news

# WebSocket (needs wscat or similar)
wscat -c "ws://localhost:8080/ws?topic=news"
```

## Architecture

All transports share a single `*broker.Broker`. Publishing via any transport
fans out to every subscriber regardless of how they're connected:

```
POST /publish ─────┐
gRPC Publish ──────┤
                   ▼
             broker.Publish(topic, payload)
                   │
       ┌───────────┼───────────┬───────────┐
       ▼           ▼           ▼           ▼
  short poll   long poll      SSE      WebSocket
  (store)      (unblock)   (flush)    (WriteFrame)
                                          │
                                    gRPC Subscribe
                                    gRPC Chat
```

## Dependency policy

- Core (broker, store, short poll, long poll, SSE, WebSocket): **stdlib only**
- gRPC: requires `google.golang.org/grpc` + `google.golang.org/protobuf`
  (gRPC framing and protobuf encoding cannot be implemented in stdlib)