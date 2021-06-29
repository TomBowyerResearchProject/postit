package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"postit/internal/api"
	"syscall"
	"time"

	"github.com/TomBowyerResearchProject/common/logger"
	"github.com/TomBowyerResearchProject/common/middlewares"
	commonPostgres "github.com/TomBowyerResearchProject/common/postgres"
	"github.com/TomBowyerResearchProject/common/verification"
)

const timeBeforeTimeout = 15

func main() {
	logger.InitLogger("postit")

	verification.Init(verification.VerificationConfig{
		VerificationURL: os.Getenv("VERIFICATION_URL"),
	})

	middlewares.Init(middlewares.Config{
		AllowedOrigin:  "*",
		AllowedMethods: "GET,POST,DELETE,OPTIONS",
		AllowedHeaders: "*",
	})

	err := commonPostgres.Connect(commonPostgres.Config{
		URI: os.Getenv("DATABASE_URL"),
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	router := api.CreateRouter()

	srv := http.Server{
		Handler:      router,
		Addr:         os.Getenv("HOST") + ":" + os.Getenv("PORT"),
		WriteTimeout: timeBeforeTimeout * time.Second,
		ReadTimeout:  timeBeforeTimeout * time.Second,
	}

	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)

		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)

		<-sigint

		logger.Infof("Shutting down server")

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			logger.Infof("HTTP server Shutdown: %v", err)
		}

		commonPostgres.CloseDatabase()

		logger.Infof("Postgres disconnected")

		close(idleConnsClosed)
	}()

	logger.Info("Starting Server")

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		// Error starting or closing listener:
		logger.Infof("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}
