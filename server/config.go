package server

import "os"

type Config struct {
	Addr string

	SessionCookieName string
	SessionSecret     string

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURL  string

	DatabasePath string
}

func LoadConfig() Config {
	return Config{
		Addr: ":8080",

		SessionCookieName: "frontiers_session",
		SessionSecret:     os.Getenv("SESSION_SECRET"),

		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),

		DiscordClientID:     os.Getenv("DISCORD_CLIENT_ID"),
		DiscordClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		DiscordRedirectURL:  os.Getenv("DISCORD_REDIRECT_URL"),

		DatabasePath: envOr("DATABASE_PATH", "frontiers.db"),
	}
}

func envOr(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
