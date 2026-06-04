package main

import (
	"log"
	"net/http"
	"time"

	"webGameGo/server"
)

func main() {
	config := server.LoadConfig()

	app, err := server.NewApp(config)
	if err != nil {
		log.Fatal(err)
	}

	if err := server.CleanupInactiveGuests(app.DB, 7*24*time.Hour); err != nil {
		log.Println("guest cleanup failed:", err)
	}

	log.Println("server listening on", config.Addr)

	if err := http.ListenAndServe(config.Addr, app.Routes()); err != nil {
		log.Fatal(err)
	}
}
