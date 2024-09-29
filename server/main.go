package main

import (
	"family-tree/server/data"
	"family-tree/server/database"
	"family-tree/server/logger"
	"family-tree/server/ui"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Failed to load .env")
	}
	if os.Getenv("PORT") == "" {
		log.Fatal("Missing ENV VAR: PORT")
	}

	logger, logfile := logger.InitLogger("./server.log")
	defer logfile.Close()

	db := database.InitNeo4jDriver()
	defer db.Driver.Close(db.Context())

	port := os.Getenv("PORT")
	logger.Info("Webserver Started @ Port: " + port)
	defer logger.Info("Webserver Ended!")

	mux := http.NewServeMux()
	dataProvider := &data.Provider{HttpHandler: mux, LogHandle: logger, DatabaseHandle: db}
	mux.HandleFunc("/data/get-person", data.MakeHandler(dataProvider, data.HelloNeo4j))

	uiProvider := &ui.Provider{HttpHandler: mux, LogHandle: logger}
	mux.HandleFunc("/", ui.MakeHandler(uiProvider, ui.HelloWorld))
	log.Fatal(http.ListenAndServe(":"+port, dataProvider))
}
