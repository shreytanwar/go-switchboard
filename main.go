package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/shreytanwar/go-switchboard/broker"
	"github.com/shreytanwar/go-switchboard/server"
	"github.com/shreytanwar/go-switchboard/store"
)

func main() {
	b := broker.New()
	st := store.New()

	httpServer := server.New(b, st)
	go func() {
		log.Println("HTTP server listening on :8080")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()


	// -------------------------------------------------------------------------
	// gRPC server — Publish (unary), Subscribe (server-stream), Chat (bidi)
	//
	// Requires the generated protobuf files and the grpc build tag:
	//   go build -tags grpc ./...
	// -------------------------------------------------------------------------

	go func() {
		lis, err := net.Listen("tcp", ":9090")
		if err != nil {
			log.Fatalf("grpc: listen: %v", err)
		}

		grpcpkg.StartGRPCServer(b, lis)
	}

	// -------------------------------------------------------------------------
	// Graceful shutdown on SIGINT / SIGTERM
	// -------------------------------------------------------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	httpServer.Close()

}