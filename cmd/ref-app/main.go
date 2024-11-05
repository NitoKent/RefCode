package main

import (
	"log/slog"
	"net/http"
	"os"

	"RefCode.com/m/internal/config"
	info "RefCode.com/m/internal/handlers/Info"
	"RefCode.com/m/internal/handlers/referral"
	"RefCode.com/m/internal/handlers/user"
	"RefCode.com/m/internal/lib/logger/sl"
	"RefCode.com/m/internal/middleware"
	"RefCode.com/m/internal/storage/sqlite"
	"github.com/gorilla/mux"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// Init Config
	cfg := config.MustLoad()

	// Init Logger
	log := setupLogger(cfg.Env)

	log.Info("Starting RefApp", slog.String("env", cfg.Env))
	log.Debug("Debug msg are enabled")

	// Init Storage
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	userHandler := user.NewHandler(storage)
	referralHandler := referral.NewHandler(storage)
	infoHandler := info.NewHandler(storage)

	// Init Router
	r := mux.NewRouter()
	// Auth
	r.HandleFunc("/register", userHandler.HandlerRegister).Methods("POST")
	r.HandleFunc("/login", userHandler.HandlerLogin).Methods("POST")
	// Ref_Code
	r.Handle("/create-referral-code", middleware.AuthMiddleware(http.HandlerFunc(referralHandler.CreateRefCode))).Methods("POST")
	r.Handle("/referral-code/delete", middleware.AuthMiddleware(http.HandlerFunc(referralHandler.DeleteRefCode))).Methods("DELETE")

	// New Handlers

	r.HandleFunc("/referral-code/email", infoHandler.GetRefCodeByEmail).Methods("POST")
	r.HandleFunc("/referrals", infoHandler.GetReferralsByIdReferrer).Methods("POST")

	// Start Server
	log.Info("Starting server on :8077")
	if err := http.ListenAndServe(":8077", r); err != nil {
		log.Error("failed to start server", sl.Err(err))
		os.Exit(1)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
