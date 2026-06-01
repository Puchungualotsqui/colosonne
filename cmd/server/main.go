package main

import (
	"log"
	"net/http"

	"webGameGo/server"
)

func main() {
	config := server.LoadConfig()

	app, err := server.NewApp(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("server listening on", config.Addr)

	if err := http.ListenAndServe(config.Addr, app.Routes()); err != nil {
		log.Fatal(err)
	}
}
