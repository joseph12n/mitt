// Command mitt is the PC hub entrypoint.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"mitt/internal/api"
	"mitt/internal/store"
	"mitt/internal/tui"
	"mitt/internal/ui"
)

// resolveToken applies the pairing token precedence: -token flag first,
// then the MITT_TOKEN environment variable, then a random token generated
// for this run. It reports whether the token was generated.
func resolveToken(flagToken string) (string, bool, error) {
	if flagToken != "" {
		return flagToken, false, nil
	}
	if env := os.Getenv("MITT_TOKEN"); env != "" {
		return env, false, nil
	}
	token, err := api.GenerateToken()
	if err != nil {
		return "", false, err
	}
	return token, true, nil
}

// maskToken returns a first-4-chars hint for logs so the full token never
// lands in log files.
func maskToken(token string) string {
	if len(token) <= 4 {
		return "…"
	}
	return token[:4] + "…"
}

// defaultDBPath returns mitt.db next to the running binary.
func defaultDBPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "mitt.db"
	}
	return filepath.Join(filepath.Dir(exe), "mitt.db")
}

func main() {
	dbFlag := flag.String("db", "", "path to sqlite database (default mitt.db next to binary)")
	addrFlag := flag.String("addr", ":8080", "LAN listen address")
	tokenFlag := flag.String("token", "", "pairing token (default MITT_TOKEN env or random)")
	uiFlag := flag.String("ui", "none", "UI mode: none (serve LAN API) or snapshot (print text dashboard and exit)")
	flag.Parse()

	dbPath := *dbFlag
	if dbPath == "" {
		dbPath = defaultDBPath()
	}

	s, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer s.Close()

	switch *uiFlag {
	case "none":
	case "snapshot":
		data, err := ui.New(s).Snapshot(context.Background())
		if err != nil {
			log.Fatalf("snapshot dashboard: %v", err)
		}
		tui.Render(os.Stdout, data)
		return
	default:
		log.Fatalf("unknown -ui mode %q: want none|snapshot", *uiFlag)
	}

	token, generated, err := resolveToken(*tokenFlag)
	if err != nil {
		log.Fatalf("resolve pairing token: %v", err)
	}
	if generated {
		// Printed once to stdout for the first pairing; never committed.
		fmt.Printf("pairing token: %s\n", token)
	}

	srv := &http.Server{
		Addr:              *addrFlag,
		Handler:           api.New(s, token),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("serve LAN API: %v", err)
		}
	}()
	log.Printf("mitt PC hub ready on %s (token %s)", *addrFlag, maskToken(token))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down mitt PC hub")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown: %v", err)
	}
}
