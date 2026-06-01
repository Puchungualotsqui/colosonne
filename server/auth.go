package server

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type SessionInfo struct {
	Token     string    `json:"-"`
	UserID    *int64    `json:"userId,omitempty"`
	GuestName string    `json:"guestName,omitempty"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type MeResponse struct {
	Authenticated bool   `json:"authenticated"`
	IsGuest       bool   `json:"isGuest"`
	UserID        *int64 `json:"userId,omitempty"`
	DisplayName   string `json:"displayName"`
	AvatarURL     string `json:"avatarUrl,omitempty"`
	Karma         int    `json:"karma"`
}

func (a *App) HandleGuestLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	session, err := a.currentSession(r)
	if err == nil && session.Token != "" {
		// Already has a valid session.
		me, err := a.meFromSession(session)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, me)
		return
	}

	guestName := "Guest" + randomDigits(4)

	session, err = a.createGuestSession(guestName, 14*24*time.Hour)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	a.setSessionCookie(w, session.Token, session.ExpiresAt)

	writeJSON(w, http.StatusOK, MeResponse{
		Authenticated: true,
		IsGuest:       true,
		DisplayName:   guestName,
		Karma:         0,
	})
}

func (a *App) HandleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	session, err := a.currentSession(r)
	if err != nil {
		writeJSON(w, http.StatusOK, MeResponse{
			Authenticated: false,
			IsGuest:       true,
			DisplayName:   "",
			Karma:         0,
		})
		return
	}

	me, err := a.meFromSession(session)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, me)
}

func (a *App) currentSession(r *http.Request) (*SessionInfo, error) {
	cookie, err := r.Cookie(a.Config.SessionCookieName)
	if err != nil {
		return nil, err
	}

	token := strings.TrimSpace(cookie.Value)
	if token == "" {
		return nil, errors.New("empty session token")
	}

	var session SessionInfo

	var userID sql.NullInt64
	var guestName sql.NullString

	err = a.DB.QueryRow(`
		SELECT token, user_id, guest_name, expires_at
		FROM sessions
		WHERE token = ?
	`, token).Scan(
		&session.Token,
		&userID,
		&guestName,
		&session.ExpiresAt,
	)

	if err != nil {
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		_, _ = a.DB.Exec(`DELETE FROM sessions WHERE token = ?`, token)
		return nil, errors.New("session expired")
	}

	if userID.Valid {
		session.UserID = &userID.Int64
	}

	if guestName.Valid {
		session.GuestName = guestName.String
	}

	return &session, nil
}

func (a *App) meFromSession(session *SessionInfo) (MeResponse, error) {
	if session.UserID == nil {
		return MeResponse{
			Authenticated: true,
			IsGuest:       true,
			DisplayName:   session.GuestName,
			Karma:         50,
		}, nil
	}

	var displayName string
	var avatarURL sql.NullString
	var karma int

	err := a.DB.QueryRow(`
		SELECT display_name, avatar_url, karma
		FROM users
		WHERE id = ?
	`, *session.UserID).Scan(&displayName, &avatarURL, &karma)

	if err != nil {
		return MeResponse{}, err
	}

	resp := MeResponse{
		Authenticated: true,
		IsGuest:       false,
		UserID:        session.UserID,
		DisplayName:   displayName,
		Karma:         karma,
	}

	if avatarURL.Valid {
		resp.AvatarURL = avatarURL.String
	}

	return resp, nil
}

func (a *App) createGuestSession(guestName string, ttl time.Duration) (*SessionInfo, error) {
	token, err := secureToken(32)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(ttl)

	_, err = a.DB.Exec(`
		INSERT INTO sessions (token, user_id, guest_name, expires_at)
		VALUES (?, NULL, ?, ?)
	`, token, guestName, expiresAt)

	if err != nil {
		return nil, err
	}

	return &SessionInfo{
		Token:     token,
		UserID:    nil,
		GuestName: guestName,
		ExpiresAt: expiresAt,
	}, nil
}

func (a *App) createUserSession(userID int64, ttl time.Duration) (*SessionInfo, error) {
	token, err := secureToken(32)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(ttl)

	_, err = a.DB.Exec(`
		INSERT INTO sessions (token, user_id, guest_name, expires_at)
		VALUES (?, ?, NULL, ?)
	`, token, userID, expiresAt)

	if err != nil {
		return nil, err
	}

	return &SessionInfo{
		Token:     token,
		UserID:    &userID,
		GuestName: "",
		ExpiresAt: expiresAt,
	}, nil
}

func (a *App) setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.Config.SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false, // Set true in production behind HTTPS.
		SameSite: http.SameSiteLaxMode,
	})
}

func secureToken(byteLen int) (string, error) {
	buf := make([]byte, byteLen)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func randomDigits(n int) string {
	buf := make([]byte, n)

	if _, err := rand.Read(buf); err != nil {
		return "0000"
	}

	out := make([]byte, n)
	for i := range buf {
		out[i] = byte('0' + int(buf[i])%10)
	}

	return string(out)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"error": message,
	})
}

// TODO: Google and Discord auth

func (a *App) HandleGoogleStart(w http.ResponseWriter, r *http.Request) {
	writeJSONError(w, http.StatusNotImplemented, "google auth not implemented yet")
}

func (a *App) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	writeJSONError(w, http.StatusNotImplemented, "google auth callback not implemented yet")
}

func (a *App) HandleDiscordStart(w http.ResponseWriter, r *http.Request) {
	writeJSONError(w, http.StatusNotImplemented, "discord auth not implemented yet")
}

func (a *App) HandleDiscordCallback(w http.ResponseWriter, r *http.Request) {
	writeJSONError(w, http.StatusNotImplemented, "discord auth callback not implemented yet")
}
