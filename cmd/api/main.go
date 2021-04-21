package main

import (
	"log"
	"net/http"
	"os"
	"postit/internal/api"
	"postit/internal/db"
	"postit/internal/event"

	"github.com/TomBowyerResearchProject/common/logger"
	"github.com/TomBowyerResearchProject/common/verification"

	"github.com/joho/godotenv"
)

func main() {
	logger.InitLogger("postit")

	verification.Init(verification.VerificationConfig{
		VerificationURL: "http://uacl/authorize",
	})

	event.Init()

	db.ConnectDB()

	router := api.CreateRouter()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")

	log.Fatal(http.ListenAndServe(host+":"+port, router))
}
