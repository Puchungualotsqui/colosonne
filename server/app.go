package server

import (
	"database/sql"
	"net/http"
)

type App struct {
	Config Config
	DB     *sql.DB
	Rooms  *RoomManager
	WS     *WebSocketServer
}

func NewApp(config Config) (*App, error) {
	db, err := OpenDB(config.DatabasePath)
	if err != nil {
		return nil, err
	}

	rooms := NewRoomManager()

	app := &App{
		Config: config,
		DB:     db,
		Rooms:  rooms,
	}

	app.WS = &WebSocketServer{
		App:   app,
		Rooms: rooms,
	}

	return app, nil
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/ws", a.WS.HandleWS)

	mux.HandleFunc("/auth/guest", a.HandleGuestLogin)
	mux.HandleFunc("/auth/google/start", a.HandleGoogleStart)
	mux.HandleFunc("/auth/google/callback", a.HandleGoogleCallback)
	mux.HandleFunc("/auth/discord/start", a.HandleDiscordStart)
	mux.HandleFunc("/auth/discord/callback", a.HandleDiscordCallback)
	mux.HandleFunc("/me", a.HandleMe)

	return mux
}
