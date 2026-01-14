package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"

	"github.com/ZaneH/keep-talking-tui/internal/tui"
)

func main() {
	const (
		host       = "0.0.0.0"
		defaultSSH = "2222"
		defaultRPC = "localhost:50051"
	)

	sshPort := getEnvOrDefault("TUI_SSH_PORT", defaultSSH)
	grpcAddr := getEnvOrDefault("TUI_GRPC_ADDR", defaultRPC)

	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%s", host, sshPort)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.MiddlewareWithProgramHandler(tui.NewProgramHandler(grpcAddr), termenv.TrueColor),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("SSH server listening on %s:%s", host, sshPort)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-done
	log.Println("shutting down server...")
	if err := s.Shutdown(context.Background()); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
