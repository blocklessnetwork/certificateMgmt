package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"certificatemgmt/internal/assign"
	"certificatemgmt/internal/config"
	"certificatemgmt/internal/middleware"
	"certificatemgmt/internal/store"

	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	client, err := store.Connect(ctx, cfg.MongoURI, cfg.MongoTimeout)
	if err != nil {
		log.Fatalf("mongodb connect: %v", err)
	}
	defer func() {
		shCtx, cancel := context.WithTimeout(context.Background(), cfg.MongoTimeout)
		defer cancel()
		if err := client.Disconnect(shCtx); err != nil {
			log.Printf("mongodb disconnect: %v", err)
		}
	}()

	if err := store.Ping(ctx, client, cfg.MongoTimeout); err != nil {
		log.Fatalf("mongodb ping: %v", err)
	}

	db := client.Database(cfg.MongoDBName)
	if err := store.EnsureIndexes(ctx, db); err != nil {
		log.Fatalf("mongodb indexes: %v", err)
	}

	mux := http.NewServeMux()
	registerRoutes(mux, db)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           middleware.RequestLog(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("http listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown: %v", err)
	}
}

func registerRoutes(mux *http.ServeMux, db *mongo.Database) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := db.Client().Ping(ctx, nil); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"status": "not_ready",
				"error":  err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"service": "certificatemgmt",
			"hints":   "GET /healthz, GET /readyz, POST /v1/assign",
		})
	})

	mux.HandleFunc("POST /v1/assign", func(w http.ResponseWriter, r *http.Request) {
		const maxBody = 1 << 20
		r.Body = http.MaxBytesReader(w, r.Body, maxBody)
		var req struct {
			UID string `json:"uid"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
			return
		}
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body"})
			return
		}
		req.UID = strings.TrimSpace(req.UID)

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		acc, err := assign.AccountByUID(ctx, db, req.UID)
		if errors.Is(err, assign.ErrUIDRequired) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, assign.ErrNoAccountAvailable) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, assign.ErrNoAssignableAccounts) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, assign.ErrAssignedAccountMissing) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		if err != nil {
			log.Printf("assign: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "assign failed"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"uid":      req.UID,
			"account":  acc.Account,
			"password": acc.Password,
			"twoFASec": acc.TwoFASec,
		})
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
